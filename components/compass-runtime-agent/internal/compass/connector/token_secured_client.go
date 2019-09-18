package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	TokenHeader = "Connector-Token"
)

//go:generate mockery -name TokenSecuredClient
type TokenSecuredClient interface {
	Configuration(token string) (schema.Configuration, error)
	SignCSR(csr string, token string) (schema.CertificationResult, error)
}

type tokenSecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewConnectorClient(endpoint string, insecureConnectorCommunication bool) TokenSecuredClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureConnectorCommunication,
			},
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &tokenSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c *tokenSecuredClient) Configuration(token string) (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c *tokenSecuredClient) SignCSR(csr string, token string) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)
	req.Header.Add(TokenHeader, token)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
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
