package director

import (
	"fmt"
	"github.com/docker/cli/cli/command/config"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/oauth"
	gcli "github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/third_party/machinebox/graphql"
	gql "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	AuthorizationHeader = "Authorization"
	TenantHeader        = "Tenant"
)

//go:generate mockery -name=DirectorClient
type DirectorClient interface {
	RegisterApplication(appName, scenario string) (string, error)
	UnregisterApplication(id string) error
	//	RequestOneTimeTokenForApplication() error
}

type directorClient struct {
	gqlClient     gql.Client
	queryProvider queryProvider
	graphqlizer   graphqlizer.Graphqlizer
	token         oauth.Token
	oauthClient   oauth.Client
}

func NewDirectorClient(gqlClient gql.Client, oauthClient oauth.Client) DirectorClient {
	return &directorClient{
		gqlClient:     gqlClient,
		oauthClient:   oauthClient,
		queryProvider: queryProvider{},
		graphqlizer:   graphqlizer.Graphqlizer{},
		token:         oauth.Token{},
	}
}

func (cc *directorClient) RegisterApplication(appName, scenario string) (string, error) {
	log.Infof("Registering Application on Director service")

	runtimeInput, err := cc.graphqlizer.RuntimeInputToGQL(directorInput)
	if err != nil {
		log.Infof("Failed to create graphQLized Runtime input")
		return "", fmt.Errorf("Failed to create graphQLized Runtime input: %s", err.Error())
	}

	registerAppQuery := cc.queryProvider.registerApplicationMutation(runtimeInput)

	appErr := cc.executeDirectorGraphQLCall(registerAppQuery, tenant, &response)
	if appErr != nil {
		return "", errors.Wrap("Failed to register runtime in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", apperrors.Internal("Failed to register runtime in Director: Received nil response.").SetComponent(apperrors.ErrCompassDirector).SetReason(apperrors.ErrDirectorNilResponse)
	}

	log.Infof("Successfully registered Runtime %s in Director for tenant %s", config.Name, tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) UnregisterApplication(appID string) error {

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
			return mapDirectorErrorToProvisionerError(egErr).Append("Failed to execute GraphQL request to Director")
		}
		return apperrors.Internal("Failed to execute GraphQL request to Director: %v", err)
	}

	return nil
}
