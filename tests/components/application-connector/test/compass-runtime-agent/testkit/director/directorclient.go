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
	RegisterApplication(appName, displayName string) (string, error)
	AssignApplicationToFormation(appId, formationName string) error
	UnregisterApplication(id string) error
	RegisterRuntime(runtimeName string) (string, error)
	UnregisterRuntime(id string) error
	RegisterFormation(formationName string) error
	UnregisterFormation(formationName string) error
	AssignRuntimeToFormation(runtimeId, formationName string) error
	GetConnectionToken(runtimeID string) (string, string, error)
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

func (cc *directorClient) RegisterFormation(formationName string) error {
	log.Infof("Registering Formation on Director service")

	registerFormationQuery := cc.queryProvider.createFormation(formationName)

	var response CreateFormationResponse
	appErr := cc.executeDirectorGraphQLCall(registerFormationQuery, cc.tenant, &response)
	if appErr != nil {
		return errors.Wrap(appErr, "Failed to register formation in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.New("Failed to register formation in Director: Received nil response.")
	}

	log.Infof("Successfully registered Formation %s in Director for tenant %s", formationName, cc.tenant)

	return nil
}

func (cc *directorClient) UnregisterFormation(formationName string) error {
	log.Infof("Unregistering Formation in Director service")

	deleteFormationQuery := cc.queryProvider.deleteFormation(formationName)

	var response DeleteFormationResponse
	appErr := cc.executeDirectorGraphQLCall(deleteFormationQuery, cc.tenant, &response)
	if appErr != nil {
		return errors.Wrap(appErr, "Failed to unregister formation in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.New("Failed to unregister formation in Director: Received nil response.")
	}

	log.Infof("Successfully unregistered Formation %s in Director for tenant %s", formationName, cc.tenant)

	return nil
}

func (cc *directorClient) RegisterRuntime(runtimeName string) (string, error) {
	log.Infof("Registering Runtime on Director service")

	registerRuntimeQuery := cc.queryProvider.registerRuntimeMutation(runtimeName)

	var response CreateRuntimeResponse
	appErr := cc.executeDirectorGraphQLCall(registerRuntimeQuery, cc.tenant, &response)
	if appErr != nil {
		return "", errors.Wrap(appErr, "Failed to register runtime in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return "", errors.New("Failed to register application in Director: Received nil response.")
	}

	log.Infof("Successfully registered Runtime %s in Director for tenant %s", runtimeName, cc.tenant)

	return response.Result.ID, nil
}

func (cc *directorClient) UnregisterRuntime(id string) error {
	log.Infof("Unregistering Runtime on Director service")
	runtimeQuery := cc.queryProvider.deleteRuntimeMutation(id)

	var response DeleteRuntimeResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, cc.tenant, &response)
	if err != nil {
		return errors.Wrap(err, "Failed to unregister runtime in Director. Request failed")
	}
	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.New("Failed to unregister runtime in Director: Received nil response.")
	}

	if response.Result.ID != id {
		return fmt.Errorf("Failed to unregister runtime %s in Director: received unexpected RuntimeID.", id)
	}

	log.Infof("Successfully unregistered Runtime %s in Director for tenant %s", id, cc.tenant)

	return nil
}

func (cc *directorClient) GetConnectionToken(runtimeId string) (string, string, error) {
	log.Infof("Requesting one time token for Runtime from Director service")
	runtimeQuery := cc.queryProvider.requestOneTimeTokenMutation(runtimeId)

	var response OneTimeTokenResponse
	err := cc.executeDirectorGraphQLCall(runtimeQuery, cc.tenant, &response)
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to get OneTimeToken for Runtime in Director. Request failed")
	}

	if response.Result == nil {
		return "", "", fmt.Errorf("Failed to get OneTimeToken for Runtime %s in Director: received nil response.", runtimeId)
	}

	log.Infof("Received OneTimeToken for Runtime %s in Director for tenant %s", runtimeId, cc.tenant)

	return response.Result.Token, response.Result.ConnectorURL, nil
}

func (cc *directorClient) RegisterApplication(appName, displayName string) (string, error) {
	log.Infof("Registering Application on Director service")
	registerAppQuery := cc.queryProvider.registerApplicationFromTemplateMutation(appName, displayName)

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

func (cc *directorClient) AssignApplicationToFormation(appId, formationName string) error {
	log.Infof("Registering Application on Director service")
	assignFormationQuery := cc.queryProvider.assignFormationForAppMutation(appId, formationName)

	var response AssignFormationResponse
	appErr := cc.executeDirectorGraphQLCall(assignFormationQuery, cc.tenant, &response)

	if appErr != nil {
		return errors.Wrap(appErr, "Failed to assign application to Formation in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.New("Failed to assign application to Formation in Director: Received nil response.")
	}

	log.Infof("Successfully assigned application %s to Formation %s in Director for tenant %s", appId, formationName, cc.tenant)

	return nil
}

func (cc *directorClient) AssignRuntimeToFormation(runtimeId, formationName string) error {
	log.Infof("Registering Application on Director service")
	assignFormationQuery := cc.queryProvider.assignFormationForRuntimeMutation(runtimeId, formationName)

	var response AssignFormationResponse
	appErr := cc.executeDirectorGraphQLCall(assignFormationQuery, cc.tenant, &response)

	if appErr != nil {
		return errors.Wrap(appErr, "Failed to assign application to Formation in Director. Request failed")
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return errors.New("Failed to assign runtime to Formation in Director: Received nil response.")
	}

	log.Infof("Successfully assigned runtime %s to Formation %s in Director for tenant %s", runtimeId, formationName, cc.tenant)

	return nil
}

func (cc *directorClient) UnregisterApplication(appID string) error {
	applicationQuery := cc.queryProvider.unregisterApplicationMutation(appID)

	var response DeleteApplicationResponse
	err := cc.executeDirectorGraphQLCall(applicationQuery, cc.tenant, &response)
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
