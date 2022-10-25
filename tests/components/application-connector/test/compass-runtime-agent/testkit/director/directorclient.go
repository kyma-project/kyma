package director

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	UnregisterApplication(id string) error
	AssignApplicationToFormation(appId, formationName string) error
	UnassignApplication(appId, formationName string) error
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
	log.Infof("Registering Formation")

	queryFunc := func() string { return cc.queryProvider.createFormation(formationName) }
	execFunc := getExecGraphQLFunc[graphql.Formation](cc)
	operationDescription := "register Formation"
	successfulLogMessage := fmt.Sprintf("Successfully registered Formation %s in Director for tenant %s", formationName, cc.tenant)

	return executeQuerySkipResponse(queryFunc, execFunc, operationDescription, successfulLogMessage)
}

func (cc *directorClient) UnregisterFormation(formationName string) error {
	log.Infof("Unregistering Formation")
	queryFunc := func() string { return cc.queryProvider.deleteFormation(formationName) }
	execFunc := getExecGraphQLFunc[graphql.Formation](cc)
	operationDescription := "unregister Formation"
	successfulLogMessage := fmt.Sprintf("Successfully unregistered Formation %s in Director for tenant %s", formationName, cc.tenant)

	return executeQuerySkipResponse(queryFunc, execFunc, operationDescription, successfulLogMessage)
}

func (cc *directorClient) RegisterRuntime(runtimeName string) (string, error) {
	log.Infof("Registering Runtime")
	queryFunc := func() string { return cc.queryProvider.registerRuntimeMutation(runtimeName) }
	execFunc := getExecGraphQLFunc[graphql.Runtime](cc)
	operationDescription := "register Runtime"
	successfulLogMessage := fmt.Sprintf("Successfully registered Runtime %s in Director for tenant %s", runtimeName, cc.tenant)

	response, err := executeQuery(queryFunc, execFunc, operationDescription, successfulLogMessage)
	if err != nil {
		return "", err
	}

	return response.Result.ID, nil
}

func (cc *directorClient) UnregisterRuntime(id string) error {
	log.Infof("Unregistering Runtime")

	queryFunc := func() string { return cc.queryProvider.deleteRuntimeMutation(id) }
	execFunc := getExecGraphQLFunc[graphql.Runtime](cc)
	operationDescription := "unregister Runtime"
	successfulLogMessage := fmt.Sprintf("Successfully unregistered Runtime %s in Director for tenant %s", id, cc.tenant)

	response, err := executeQuery(queryFunc, execFunc, operationDescription, successfulLogMessage)
	if err != nil {
		return err
	}

	if response.Result.ID != id {
		return fmt.Errorf("Failed to unregister runtime %s in Director: received unexpected RuntimeID.", id)
	}

	return nil
}

func (cc *directorClient) GetConnectionToken(runtimeId string) (string, string, error) {
	log.Infof("Requesting one time token for Runtime from Director service")

	queryFunc := func() string { return cc.queryProvider.requestOneTimeTokenMutation(runtimeId) }
	execFunc := getExecGraphQLFunc[graphql.OneTimeTokenForRuntimeExt](cc)
	operationDescription := "register application"
	successfulLogMessage := fmt.Sprintf("Received OneTimeToken for Runtime %s in Director for tenant %s", runtimeId, cc.tenant)

	response, err := executeQuery(queryFunc, execFunc, operationDescription, successfulLogMessage)
	if err != nil {
		return "", "", err
	}
	return response.Result.Token, response.Result.ConnectorURL, nil
}

func (cc *directorClient) RegisterApplication(appName, displayName string) (string, error) {
	log.Infof("Registering Application")

	queryFunc := func() string { return cc.queryProvider.registerApplicationFromTemplateMutation(appName, displayName) }
	execFunc := getExecGraphQLFunc[graphql.Application](cc)
	operationDescription := "register application"
	successfulLogMessage := fmt.Sprintf("Successfully registered application %s in Director for tenant %s", appName, cc.tenant)

	result, err := executeQuery(queryFunc, execFunc, operationDescription, successfulLogMessage)
	if err != nil {
		return "", err
	}
	return result.Result.ID, err
}

func (cc *directorClient) AssignApplicationToFormation(appId, formationName string) error {
	log.Infof("Assigning Application to Formation")

	queryFunc := func() string { return cc.queryProvider.assignFormationForAppMutation(appId, formationName) }
	execFunc := getExecGraphQLFunc[graphql.Formation](cc)
	operationDescription := "assign Application to Formation"
	successfulLogMessage := fmt.Sprintf("Successfully assigned application %s to Formation %s in Director for tenant %s", appId, formationName, cc.tenant)

	return executeQuerySkipResponse(queryFunc, execFunc, operationDescription, successfulLogMessage)
}

func (cc *directorClient) UnassignApplication(appId, formationName string) error {
	log.Infof("Unregistering Application from Formation")

	queryFunc := func() string { return cc.queryProvider.unassignFormation(appId, formationName) }
	execFunc := getExecGraphQLFunc[graphql.Formation](cc)
	operationDescription := "unregister formation"
	successfulLogMessage := fmt.Sprintf("Successfully unassigned application %s from Formation %s in Director for tenant %s", appId, formationName, cc.tenant)

	return executeQuerySkipResponse(queryFunc, execFunc, operationDescription, successfulLogMessage)
}

func (cc *directorClient) AssignRuntimeToFormation(runtimeId, formationName string) error {
	log.Infof("Assigning Runtime to Formation")

	queryFunc := func() string { return cc.queryProvider.assignFormationForRuntimeMutation(runtimeId, formationName) }
	execFunc := getExecGraphQLFunc[graphql.Formation](cc)
	operationDescription := "assign Runtime to Formation"
	successfulLogMessage := fmt.Sprintf("Successfully assigned runtime %s to Formation %s in Director for tenant %s", runtimeId, formationName, cc.tenant)

	return executeQuerySkipResponse(queryFunc, execFunc, operationDescription, successfulLogMessage)
}

func (cc *directorClient) UnregisterApplication(appID string) error {
	log.Infof("Unregistering Application")

	queryFunc := func() string { return cc.queryProvider.unregisterApplicationMutation(appID) }
	execFunc := getExecGraphQLFunc[graphql.Application](cc)
	operationDescription := "Unregistering Application"
	successfulLogMessage := fmt.Sprintf("Successfully unregister application %s in Director for tenant %s", appID, cc.tenant)

	response, err := executeQuery(queryFunc, execFunc, operationDescription, successfulLogMessage)
	if err != nil {
		return err
	}

	if response.Result.ID != appID {
		return fmt.Errorf("Failed to unregister Application %s in Director: received unexpected applicationID.", appID)
	}

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

type Response[T any] struct {
	Result *T
}

func executeQuerySkipResponse[T any](getQueryFunc func() string, executeQueryFunc func(string, *Response[T]) error, operationDescription, successfulLogMessage string) error {
	_, err := executeQuery(getQueryFunc, executeQueryFunc, operationDescription, successfulLogMessage)

	return err
}

func executeQuery[T any](getQueryFunc func() string, executeQueryFunc func(string, *Response[T]) error, operationDescription, successfulLogMessage string) (Response[T], error) {
	query := getQueryFunc()

	var response Response[T]
	err := executeQueryFunc(query, &response)

	if err != nil {
		return Response[T]{}, errors.Wrap(err, fmt.Sprintf("Failed to %s in Director. Request failed", operationDescription))
	}

	// Nil check is necessary due to GraphQL client not checking response code
	if response.Result == nil {
		return Response[T]{}, errors.New(fmt.Sprintf("Failed to %s in Director: Received nil response.", operationDescription))
	}

	log.Infof(successfulLogMessage)

	return response, nil
}

func getExecGraphQLFunc[T any](cc *directorClient) func(string, *Response[T]) error {
	return func(query string, result *Response[T]) error {
		if cc.token.EmptyOrExpired() {
			log.Infof("Refreshing token to access Director Service")
			if err := cc.getToken(); err != nil {
				return err
			}
		}

		req := gcli.NewRequest(query)
		req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", cc.token.AccessToken))
		req.Header.Set(TenantHeader, cc.tenant)

		if err := cc.gqlClient.Do(req, result); err != nil {
			if egErr, ok := err.(gcli.ExtendedError); ok {
				return errors.Wrap(egErr, "Failed to execute GraphQL request to Director")
			}
			return fmt.Errorf("Failed to execute GraphQL request to Director: %v", err)
		}

		return nil
	}
}
