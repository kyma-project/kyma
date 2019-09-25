package compass

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqltools "github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
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

	authorizationToken string
}

// TODO: client will need to be authenticated after implementation of certs
func NewCompassClient(endpoint, tenant, runtimeId, scenarioLabel, token string, gqlLog bool) *Client {
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
		client:             client,
		graphqlizer:        &gqltools.Graphqlizer{},
		queryProvider:      queryProvider{},
		tenant:             tenant,
		scenarioLabel:      scenarioLabel,
		runtimeId:          runtimeId,
		authorizationToken: token,
	}
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

	var response ScenarioLabelDefinitionResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to get scenarios label definition")
	}

	return response.Result.Schema, nil
}

func (c *Client) updateScenarios(schema ScenariosSchema) (ScenariosSchema, error) {
	gqlInput, err := c.graphqlizer.LabelDefinitionInputToGQL(schema.ToLabelDefinitionInput(ScenariosLabelName))
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to convert LabelDefinitionInput")
	}
	query := c.queryProvider.updateLabelDefinition(gqlInput)
	req := c.newRequest(query)

	var response ScenarioLabelDefinitionResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return ScenariosSchema{}, errors.Wrap(err, "Failed to update scenarios label definition")
	}

	return response.Result.Schema, nil
}

func (c *Client) labelRuntime(values []string) error {
	query := c.queryProvider.setRuntimeLabel(c.runtimeId, ScenariosLabelName, values)

	req := c.newRequest(query)

	var response SetRuntimeLabelResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return errors.Wrap(err, "Failed to label runtime with scenarios")
	}

	return nil
}

// Applications

func (c *Client) CreateApplication(input graphql.ApplicationInput) (Application, error) {
	c.setScenarioLabel(&input)

	appInputGQL, err := c.graphqlizer.ApplicationInputToGQL(input)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.createApplication(appInputGQL)
	req := c.newRequest(query)

	var response ApplicationResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to create Application")
	}

	return response.Result, nil
}

func (c *Client) UpdateApplication(applicationId string, input graphql.ApplicationInput) (Application, error) {
	c.setScenarioLabel(&input)

	appInputGQL, err := c.graphqlizer.ApplicationInputToGQL(input)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.updateApplication(applicationId, appInputGQL)
	req := c.newRequest(query)

	var response ApplicationResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to update Application")
	}

	return response.Result, nil
}

func (c *Client) DeleteApplication(id string) (string, error) {
	query := c.queryProvider.deleteApplication(id)

	req := c.newRequest(query)

	var response DeleteResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to delete Application")
	}

	return response.Result.ID, nil
}

func (c *Client) setScenarioLabel(input *graphql.ApplicationInput) {
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

	var response APIResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *Client) DeleteAPI(id string) (string, error) {
	query := c.queryProvider.deleteAPI(id)
	req := c.newRequest(query)

	var response DeleteResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", err
	}

	return response.Result.ID, nil
}

// Event APIs

func (c *Client) CreateEventAPI(appId string, input graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	api, err := c.modifyEventAPI(appId, input, c.queryProvider.createEventAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Event API")
	}

	return api, nil
}

func (c *Client) UpdateEventAPI(apiId string, input graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	api, err := c.modifyEventAPI(apiId, input, c.queryProvider.updateEventAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update Event API")
	}

	return api, nil
}

func (c *Client) modifyEventAPI(id string, input graphql.EventAPIDefinitionInput, prepareQuery func(applicationId string, input string) string) (*graphql.EventAPIDefinition, error) {
	eventAPIInputGQL, err := c.graphqlizer.EventAPIDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	query := prepareQuery(id, eventAPIInputGQL)
	req := c.newRequest(query)

	var response CreateEventAPIResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (c *Client) DeleteEventAPI(id string) (string, error) {
	query := c.queryProvider.deleteEventAPI(id)
	req := c.newRequest(query)

	var response DeleteResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", err
	}

	return response.Result.ID, nil
}

func (c *Client) newRequest(query string) *gcli.Request {
	req := gcli.NewRequest(query)
	req.Header.Set(TenantHeader, c.tenant)
	req.Header.Set(AuthorizationHeader, fmt.Sprintf("Bearer %s", c.authorizationToken))
	return req
}
