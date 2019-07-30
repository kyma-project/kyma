package compass

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqltools "github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TenantHeader = "Tenant"
)

type Client struct {
	client        *gcli.Client
	graphqlizer   gqltools.Graphqlizer
	queryProvider queryProvider

	tenant    string
	runtimeId string // TODO - I might not need it
}

// TODO: pass options with authenticated client
func NewCompassClient(endpoint, tenant, runtimeId string) *Client {
	client := gcli.NewClient(endpoint)
	client.Log = func(s string) {
		logrus.Info(s)
	}

	return &Client{
		client:        gcli.NewClient(endpoint),
		graphqlizer:   gqltools.Graphqlizer{},
		queryProvider: queryProvider{},
		tenant:        tenant,
		runtimeId:     runtimeId,
	}
}

// TODO - will it unmarshal correctly if I replace graphql models with my aliases?
type Application struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description *string                         `json:"description"`
	Labels      map[string][]string             `json:"labels"`
	APIs        *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs   *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents   *graphql.DocumentPage           `json:"documents"`
}

type CreateApplicationResponse struct {
	Result Application `json:"result"`
}

type CreateAPIResponse struct {
	Result *graphql.APIDefinition `json:"result"`
}

type CreateEventAPIResponse struct {
	Result *graphql.EventAPIDefinition `json:"result"`
}

type DeleteResponse struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

type IdResponse struct {
	Id string `json:"id"`
}

// Applications

func (c *Client) CreateApplication(input graphql.ApplicationInput) (Application, error) {
	appInputGQL, err := c.graphqlizer.ApplicationInputToGQL(input)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.createApplication(appInputGQL)
	req := c.newRequest(query)

	var response CreateApplicationResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return Application{}, errors.Wrap(err, "Failed to create Application")
	}

	return response.Result, nil
}

func (c *Client) UpdateApplication(input graphql.ApplicationInput) (Application, error) {
	// TODO - implement

	//appInputGQL, err := c.graphqlizer.ApplicationInputToGQL(input)
	//if err != nil {
	//	return Application{}, errors.Wrap(err, "Failed to convert Application Input to query")
	//}
	//
	//query := c.queryProvider.createApplication(appInputGQL)
	//
	//req := gcli.NewRequest(query)
	//req.Header.Set(TenantHeader, c.tenant)
	//
	//var response CreateApplicationResponse
	//err = c.client.Run(context.Background(), req, &response)
	//if err != nil {
	//	return Application{}, errors.Wrap(err, "Failed to create Application")
	//}
	//
	//return response.Result, nil

	return Application{}, nil
}

func (c *Client) DeleteApplication(id string) (string, error) {
	query := c.queryProvider.deleteApplication(id)

	req := gcli.NewRequest(query)
	req.Header.Set(TenantHeader, c.tenant)

	var response DeleteResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to delete Application")
	}

	return response.Result.ID, nil
}

// APIs

func (c *Client) CreateAPI(appId string, input graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	api, err := c.modifyAPI(appId, input, c.queryProvider.createAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create API")
	}

	return api, nil
}

func (c *Client) UpdateAPI(appId string, input graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	api, err := c.modifyAPI(appId, input, c.queryProvider.updateAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update API")
	}

	return api, nil
}

func (c *Client) modifyAPI(appId string, input graphql.APIDefinitionInput, prepareQuery func(applicationId string, input string) string) (*graphql.APIDefinition, error) {
	appInputGQL, err := c.graphqlizer.APIDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	query := prepareQuery(appId, appInputGQL)
	req := c.newRequest(query)

	var response CreateAPIResponse
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

func (c *Client) UpdateEventAPI(appId string, input graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	api, err := c.modifyEventAPI(appId, input, c.queryProvider.updateEventAPI)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to update Event API")
	}

	return api, nil
}

func (c *Client) modifyEventAPI(appId string, input graphql.EventAPIDefinitionInput, prepareQuery func(applicationId string, input string) string) (*graphql.EventAPIDefinition, error) {
	eventAPIInputGQL, err := c.graphqlizer.EventAPIDefinitionInputToGQL(input)
	if err != nil {
		return nil, err
	}

	query := prepareQuery(appId, eventAPIInputGQL)
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
	return req
}
