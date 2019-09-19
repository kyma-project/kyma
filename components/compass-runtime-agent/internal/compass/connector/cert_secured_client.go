package connector

import (
	"context"
	"crypto/tls"
	"net/http"

	schema "github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=CertificateSecuredClient
type CertificateSecuredClient interface {
	Configuration() (schema.Configuration, error)
	SignCSR(csr string) (schema.CertificationResult, error)
}

type certificateSecuredClient struct {
	graphQlClient *gcli.Client
	queryProvider queryProvider
}

func NewCertificateSecuredConnectorClient(endpoint string, skipTLSVerify bool, certificates ...tls.Certificate) CertificateSecuredClient {
	tlsConfig := &tls.Config{
		Certificates:       certificates,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: skipTLSVerify,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	graphQlClient := gcli.NewClient(endpoint, gcli.WithHTTPClient(httpClient))

	return &certificateSecuredClient{
		graphQlClient: graphQlClient,
		queryProvider: queryProvider{},
	}
}

func (c certificateSecuredClient) Configuration() (schema.Configuration, error) {
	query := c.queryProvider.configuration()
	req := gcli.NewRequest(query)

	var response ConfigurationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.Configuration{}, errors.Wrap(err, "Failed to get configuration")
	}
	return response.Result, nil
}

func (c certificateSecuredClient) SignCSR(csr string) (schema.CertificationResult, error) {
	query := c.queryProvider.signCSR(csr)
	req := gcli.NewRequest(query)

	var response CertificationResponse

	err := c.graphQlClient.Run(context.Background(), req, &response)
	if err != nil {
		return schema.CertificationResult{}, errors.Wrap(err, "Failed to generate certificate")
	}
	return response.Result, nil
}
