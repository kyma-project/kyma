package compass

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/avast/retry-go"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqltools "github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	TenantHeader        = "Tenant"
	ScenariosLabelName  = "scenarios"
	AuthorizationHeader = "Authorization"
)

type Client struct {
	client        *gcli.Client
	graphqlizer   *gqltools.Graphqlizer
	queryProvider queryProvider

	tenant        string
	runtimeId     string
	scenarioLabel string

	directorToken string
}

func NewCompassClient(endpoint, tenant, runtimeId, scenarioLabel string, gqlLog bool) *Client {

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	client := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))
	if gqlLog {
		client.Log = func(s string) {
			logrus.Info(s)
		}
	}

	return &Client{
		client:        client,
		graphqlizer:   &gqltools.Graphqlizer{},
		queryProvider: queryProvider{},
		tenant:        tenant,
		scenarioLabel: scenarioLabel,
		runtimeId:     runtimeId,
	}
}

// Runtimes

func (c *Client) GetRuntime(runtimeId string) (Runtime, error) {
	query := c.queryProvider.getRuntime(runtimeId)
	req := c.newRequest(query)

	var runtime Runtime
	err := c.executeRequest(req, &runtime, &Runtime{})
	if err != nil {
		return Runtime{}, errors.Wrap(err, "Failed to get Runtime")
	}

	return runtime, nil
}

// Setup

func (c *Client) SetDirectorToken(directorToken string) {
	c.directorToken = directorToken
}

// Scenario labels

func (c *Client) SetupTestsScenario() error {
	scenarios, err := c.getScenarios()
	if err != nil {
		return errors.Wrap(err, "Failed to setup tests scenario")
	}

	scenarios.AddScenario(c.scenarioLabel)

	scenarios, err = c.updateScenarios(scenarios)
	if err != nil {
		return errors.Wrap(err, "Failed to setup tests scenario")
	}

	return c.labelRuntime(scenarios.Items.Enum)
}

func (c *Client) CleanupTestsScenario() error {
	scenarios, err := c.getScenarios()
	if err != nil {
		return errors.Wrap(err, "Failed to cleanup tests scenario")
	}

	scenarios.RemoveScenario(c.scenarioLabel)

	err = c.labelRuntime(scenarios.Items.Enum)
	if err != nil {
		return errors.Wrap(err, "Failed to label Runtime")
	}

	scenarios, err = c.updateScenarios(scenarios)
	if err != nil {
		return errors.Wrap(err, "Failed to cleanup tests scenario")
	}

	return nil
}

func (c *Client) getScenarios() (ScenariosSchema, error) {
	query := c.queryProvider.labelDefinition(ScenariosLabelName)
	req := c.newRequest(query)

	var labelDefinition ScenarioLabelDefinition
	err := c.executeRequest(req, &labelDefinition, &ScenarioLabelDefinition{})
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to get scenarios label definition")
	}

	scenarioSchema, err := ToScenarioSchema(labelDefinition)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to get scenario schema")
	}

	return scenarioSchema, nil
}

func (c *Client) updateScenarios(schema ScenariosSchema) (ScenariosSchema, error) {
	labelDef, err := schema.ToLabelDefinitionInput(ScenariosLabelName)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to convert ScenarioSchema")
	}

	gqlInput, err := c.graphqlizer.LabelDefinitionInputToGQL(labelDef)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to convert LabelDefinitionInput")
	}
	query := c.queryProvider.updateLabelDefinition(gqlInput)
	req := c.newRequest(query)

	var labelDefinition ScenarioLabelDefinition
	err = c.executeRequest(req, &labelDefinition, &ScenarioLabelDefinition{})
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to update scenarios label definition")
	}

	scenarioSchema, err := ToScenarioSchema(labelDefinition)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to get scenario schema")
	}

	return scenarioSchema, nil
}

func (c *Client) labelRuntime(values []string) error {
	query := c.queryProvider.setRuntimeLabel(c.runtimeId, ScenariosLabelName, values)

	req := c.newRequest(query)

	var label *graphql.Label
	err := c.executeRequest(req, label, graphql.Label{})
	if err != nil {
		return errors.Wrap(err, "Failed to label runtime with scenarios")
	}

	return nil
}

// Applications

func (c *Client) GetOneTimeTokenForApplication(applicationId string) (graphql.OneTimeToken, error) {
	query := c.queryProvider.requestOneTimeTokenForApplication(applicationId)
	req := c.newRequest(query)

	var oneTimeToken graphql.OneTimeToken
	err := c.executeRequest(req, &oneTimeToken, &graphql.OneTimeToken{})
	if err != nil {
		return graphql.OneTimeToken{}, errors.Wrap(err, "Failed to update Application")
	}

	return oneTimeToken, nil
}

func (c *Client) CreateApplication(input graphql.ApplicationRegisterInput) (Application, error) {
	c.setScenarioLabel(&input)

	appInputGQL, err := c.graphqlizer.ApplicationCreateInputToGQL(input)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.createApplication(appInputGQL)
	req := c.newRequest(query)

	var application Application
	err = c.executeRequest(req, &application, &Application{})
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to create Application")
	}

	return application, nil
}

func (c *Client) UpdateApplication(applicationId string, input graphql.ApplicationUpdateInput) (Application, error) {
	appInputGQL, err := c.graphqlizer.ApplicationUpdateInputToGQL(input)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.updateApplication(applicationId, appInputGQL)
	req := c.newRequest(query)

	var application Application
	err = c.executeRequest(req, &application, &Application{})
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to update Application")
	}

	return application, nil
}

func (c *Client) DeleteApplication(id string) (string, error) {
	query := c.queryProvider.deleteApplication(id)

	req := c.newRequest(query)

	var response IdResponse
	err := c.executeRequest(req, &response, &IdResponse{})
	if err != nil {
		return "", errors.Wrap(err, "Failed to delete Application")
	}

	return response.Id, nil
}

func (c *Client) setScenarioLabel(input *graphql.ApplicationRegisterInput) {
	var labels = map[string]interface{}{
		ScenariosLabelName: []string{c.scenarioLabel},
	}

	gqlLabels := graphql.Labels(labels)
	input.Labels = &gqlLabels
}

// APIs

func (c *Client) CreateAPI(appId string, input graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	api, err := c.modifyAPI(appId, input, c.queryProvider.createAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create API")
	}

	return api, nil
}

func (c *Client) UpdateAPI(apiId string, input graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	api, err := c.modifyAPI(apiId, input, c.queryProvider.updateAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update API")
	}

	return api, nil
}

func (c *Client) modifyAPI(id string, input graphql.APIDefinitionInput, prepareQuery func(applicationId string, input string) string) (*graphql.APIDefinition, error) {
	appInputGQL, err := c.graphqlizer.APIDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	query := prepareQuery(id, appInputGQL)
	req := c.newRequest(query)

	var apiDef graphql.APIDefinition
	err = c.executeRequest(req, &apiDef, &graphql.APIDefinition{})
	if err != nil {
		return nil, err
	}

	return &apiDef, nil
}

func (c *Client) DeleteAPI(id string) (string, error) {
	query := c.queryProvider.deleteAPI(id)
	req := c.newRequest(query)

	var response IdResponse
	err := c.executeRequest(req, &response, &IdResponse{})
	if err != nil {
		return "", err
	}

	return response.Id, nil
}

// Event APIs

func (c *Client) CreateEventAPI(appId string, input graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	api, err := c.modifyEventAPI(appId, input, c.queryProvider.createEventAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Event API")
	}

	return api, nil
}

func (c *Client) UpdateEventAPI(apiId string, input graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	api, err := c.modifyEventAPI(apiId, input, c.queryProvider.updateEventAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update Event API")
	}

	return api, nil
}

func (c *Client) modifyEventAPI(id string, input graphql.EventDefinitionInput, prepareQuery func(applicationId string, input string) string) (*graphql.EventDefinition, error) {
	eventAPIInputGQL, err := c.graphqlizer.EventDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	query := prepareQuery(id, eventAPIInputGQL)
	req := c.newRequest(query)

	var eventAPIDef graphql.EventDefinition
	err = c.executeRequest(req, &eventAPIDef, &graphql.EventDefinition{})
	if err != nil {
		return nil, err
	}

	return &eventAPIDef, nil
}

func (c *Client) DeleteEventAPI(id string) (string, error) {
	query := c.queryProvider.deleteEventAPI(id)
	req := c.newRequest(query)

	var response IdResponse
	err := c.executeRequest(req, &response, &IdResponse{})
	if err != nil {
		return "", err
	}

	return response.Id, nil
}

func (c *Client) newRequest(query string) *gcli.Request {
	req := gcli.NewRequest(query)
	req.Header.Set(TenantHeader, c.tenant)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", c.directorToken))
	return req
}

func (c *Client) executeRequest(req *gcli.Request, destination interface{}, emptyObject interface{}) error {
	if reflect.ValueOf(destination).Kind() != reflect.Ptr {
		return errors.New("destination is not of pointer type")
	}

	wrapper := &graphQLResponseWrapper{Result: destination}

	return retry.Do(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := c.client.Run(ctx, req, wrapper)
		if err != nil {
			return errors.Wrap(err, "Failed to execute request")
		}

		// Due to GraphQL client not checking response codes we need to relay on result being empty in case of failure
		if reflect.DeepEqual(destination, emptyObject) {
			return errors.New("Failed to execute request: received empty object response")
		}

		return nil
	}, defaultRetryOptions()...)

}

func defaultRetryOptions() []retry.Option {
	return []retry.Option{retry.Attempts(6), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second)}
}
