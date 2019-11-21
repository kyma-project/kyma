package connector

import (
	"kyma-project.io/compass-runtime-agent/internal/graphql"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Client
type Client interface {
	Configuration(headers map[string]string) (schema.Configuration, error)
	SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error)
}

type connectorClient struct {
	graphQlClient graphql.Client
	queryProvider queryProvider
}

func NewConnectorClient(graphQlClient graphql.Client) Client {
	return &connectorClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c connectorClient) Configuration(headers map[string]string) (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)

	applyHeaders(req, headers)

	var response ConfigurationResponse

	err := c.graphQlClient.Do(req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c connectorClient) SignCSR(csr string, headers map[string]string) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)

	applyHeaders(req, headers)

	var response CertificationResponse

	err := c.graphQlClient.Do(req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}

func applyHeaders(req *gcli.Request, headers map[string]string) {
	for h, val := range headers {
		req.Header.Set(h, val)
	}
}

type ConfigurationResponse struct {
	Result schema.Configuration `json:"result"`
}

type CertificationResponse struct {
	Result schema.CertificationResult `json:"result"`
}

type RevokeResult struct {
	Result bool `json:"result"`
}
