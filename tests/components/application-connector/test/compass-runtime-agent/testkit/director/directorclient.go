package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	gql "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/graphql"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	gcli "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/third_party/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

//go:generate mockery --name=Client
type Client interface {
	RegisterApplication(appName, scenario string) (string, error)
	UnregisterApplication(id string) error
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer.Graphqlizer
	token         oauth.Token
	oauthClient   oauth.Client
	tenant        string
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client, tenant string) Client {

	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer.Graphqlizer{},
		token:         oauth.Token{},
		tenant:        tenant,
	}
}

func (cc *directorClient) getToken() error {
	token, err := cc.oauthClient.GetAuthorizationToken()
	if err != nil {
		return errors.New("Error while obtaining token")
	}

	if token.EmptyOrExpired() {
		return errors.New("Obtained empty or expired token")
	}

	cc.token = token
	return nil
}

func (cc *directorClient) RegisterApplication(appName, scenario string) (string, error) {
	log.Infof("Registering Application on Director service")

	//var labels graphql.Labels
	//if config.Labels != nil {
	//	l := graphql.Labels(config.Labels)
	//	labels = l
	//}
	//
	//
	//directorInput := graphql.ApplicationRegisterInput{
	//	Name: appName,
	//	//Labels:      labels,
	//}
	//appInput, err := cc.graphqlizer.ApplicationRegisterInputToGQL(directorInput)
	//if err != nil {
	//	log.Infof("Failed to create graphQLized Runtime input")
	//	return "", fmt.Errorf("Failed to create graphQLized Runtime input: %s", err.Error())
	//}

	registerAppQuery := cc.queryProvider.registerApplicationMutation(appName, scenario)

	var response CreateApplicationResponse
	appErr := cc.executeDirectorGraphQLCall(registerAppQuery, cc.tenant, &response)
	if appErr != nil {
		return "", errors.Wrap(appErr, "Failed to register application in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.New("Failed to register application in Director: Received nil response.")
	}

	log.Infof("Successfully registered application %s in Director for tenant %s", appName, cc.tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) UnregisterApplication(appID string) error {
	runtimeQuery := cc.queryProvider.unregisterApplicationMutation(appID)

	var response DeleteApplicationResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, cc.tenant, &response)
	if err != nil {
		return fmt.Errorf("Failed to unregister application %s in Director", appID)
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return fmt.Errorf("Failed to unregister application %s in Director: received nil response.", appID)
	}

	if response.Result.ID != appID {
		return fmt.Errorf("Failed to unregister application %s in Director: received unexpected applicationID.", appID)
	}

	log.Infof("Successfully unregistered Application %s in Director for tenant %s", appID, cc.tenant)

	return nil
}

func (cc *directorClient) executeDirectorGraphQLCall(directorQuery string, tenant string, response interface{}) error {
	if cc.token.EmptyOrExpired() {
		log.Infof("Refreshing token to access Director Service")
		if err := cc.getToken(); err != nil {
			return err
		}
	}

	req := gcli.NewRequest(directorQuery)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
	req.Header.Set(TenantHeader, tenant)

	if err := cc.gqlClient.Do(req, response); err != nil {
		if egErr, ok := err.(gcli.ExtendedError); ok {
			return errors.Wrap(egErr, "Failed to execute GraphQL request to Director")
		}
		return fmt.Errorf("Failed to execute GraphQL request to Director: %v", err)
	}

	return nil
}
