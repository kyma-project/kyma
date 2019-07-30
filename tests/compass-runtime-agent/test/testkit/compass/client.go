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
	return &Client{
		client:        gcli.NewClient(endpoint),
		graphqlizer:   gqltools.Graphqlizer{},
		queryProvider: queryProvider{},
		tenant:        tenant,
		runtimeId:     runtimeId,
	}
}

// TODO -fetch full response for asserts
type AppIdsResponse struct {
	Id           string
	APIsIds      []string
	EventAPIsIds []string
	DocumentsIds []string
}

//type ApplicationPage struct {
//	Data       []*Application    `json:"data"`
//	PageInfo   *graphql.PageInfo `json:"pageInfo"`
//	TotalCount int               `json:"totalCount"`
//}
//
//type Application struct {
//	ID             string                          `json:"id"`
//	Name           string                          `json:"name"`
//	Description    *string                         `json:"description"`
//	Labels         map[string][]string             `json:"labels"`
//	Status         *graphql.ApplicationStatus      `json:"status"`
//	Webhooks       []*graphql.Webhook              `json:"webhooks"`
//	APIs           *graphql.APIDefinitionPage      `json:"apis"`
//	EventAPIs      *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
//	Documents      *graphql.DocumentPage           `json:"documents"`
//	HealthCheckURL *string                         `json:"healthCheckURL"`
//}

type CreateApplicationResponse struct {
	Result struct {
		ID   string `json:"id"`
		APIs struct {
			Data []IdResponse `json:"data"`
		} `json:"apis"`
		EventAPIs struct {
			Data []IdResponse `json:"data"`
		} `json:"eventAPIs"`
		Documents struct {
			Data []IdResponse `json:"data"`
		} `json:"documents"`
	} `json:"result"`
}

type DeleteApplicationResponse struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

type IdResponse struct {
	Id string `json:"id"`
}

func rawIDs(idResponses []IdResponse) []string {
	ids := make([]string, len(idResponses))
	for i, idResp := range idResponses {
		ids[i] = idResp.Id
	}
	return ids
}

func (c *Client) CreateApplication(input graphql.ApplicationInput) (AppIdsResponse, error) {
	appInputGQL, err := c.graphqlizer.ApplicationInputToGQL(input)
	if err != nil {
		return AppIdsResponse{}, errors.Wrap(err, "Failed to convert Application Input to query")
	}

	query := c.queryProvider.createApplication(appInputGQL)

	req := gcli.NewRequest(query)
	req.Header.Set(TenantHeader, c.tenant)

	var response CreateApplicationResponse
	err = c.client.Run(context.Background(), req, &response)
	if err != nil {
		return AppIdsResponse{}, errors.Wrap(err, "Failed to create Application")
	}

	return AppIdsResponse{
		Id:           response.Result.ID,
		APIsIds:      rawIDs(response.Result.APIs.Data),
		EventAPIsIds: rawIDs(response.Result.EventAPIs.Data),
		DocumentsIds: rawIDs(response.Result.Documents.Data),
	}, nil
}

func (c *Client) DeleteApplication(id string) (string, error) {
	query := c.queryProvider.deleteApplication(id)

	c.client.Log = func(s string) {
		logrus.Info(s)
	}

	req := gcli.NewRequest(query)
	req.Header.Set(TenantHeader, c.tenant)

	var response DeleteApplicationResponse
	err := c.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", errors.Wrap(err, "Failed to delete Application")
	}

	return response.Result.ID, nil
}
