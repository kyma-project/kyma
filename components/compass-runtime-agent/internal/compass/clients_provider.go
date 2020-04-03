package compass

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"kyma-project.io/compass-runtime-agent/internal/compass/connector"
	"kyma-project.io/compass-runtime-agent/internal/compass/director"
	"kyma-project.io/compass-runtime-agent/internal/config"
	"kyma-project.io/compass-runtime-agent/internal/graphql"
)

//go:generate mockery -name=ClientsProvider
type ClientsProvider interface {
	GetDirectorClient(runtimeConfig config.RuntimeConfig) (director.DirectorClient, error)
	GetConnectorTokensClient(url string) (connector.Client, error)
	GetConnectorCertSecuredClient() (connector.Client, error)
}

func NewClientsProvider(gqlClientConstr graphql.ClientConstructor, skipCompassTLSVerification, enableLogging bool) *clientsProvider {
	return &clientsProvider{
		gqlClientConstructor:       gqlClientConstr,
		skipCompassTLSVerification: skipCompassTLSVerification,
		enableLogging:              enableLogging,

		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type clientsProvider struct {
	gqlClientConstructor       graphql.ClientConstructor
	skipCompassTLSVerification bool
	enableLogging              bool
	httpClient                 *http.Client

	// lazy init after establishing connection
	mtlsHTTPClient      *http.Client
	connectorSecuredURL string
	directorURL         string
}

func (cp *clientsProvider) UpdateConnectionData(data ConnectionData) error {
	var transport *http.Transport
	if cp.mtlsHTTPClient == nil {
		cp.mtlsHTTPClient = &http.Client{Timeout: 30 * time.Second}
		transport = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		transport = cp.mtlsHTTPClient.Transport.(*http.Transport)
	}

	transport.TLSClientConfig.InsecureSkipVerify = cp.skipCompassTLSVerification
	transport.TLSClientConfig.Certificates = []tls.Certificate{data.Certificate}

	cp.mtlsHTTPClient.Transport = transport

	cp.directorURL = data.DirectorURL
	cp.connectorSecuredURL = data.ConnectorURL

	return nil
}

func (cp *clientsProvider) GetDirectorClient(runtimeConfig config.RuntimeConfig) (director.DirectorClient, error) {
	if cp.mtlsHTTPClient == nil {
		return nil, fmt.Errorf("failed to get Director client: mTLS HTTP client not initialized")
	}

	gqlClient, err := cp.gqlClientConstructor(cp.mtlsHTTPClient, cp.directorURL, cp.enableLogging)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return director.NewConfigurationClient(gqlClient, runtimeConfig), nil
}

func (cp *clientsProvider) GetConnectorTokensClient(url string) (connector.Client, error) {
	gqlClient, err := cp.gqlClientConstructor(cp.httpClient, url, cp.enableLogging)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}

func (cp *clientsProvider) GetConnectorCertSecuredClient() (connector.Client, error) {
	if cp.mtlsHTTPClient == nil {
		return nil, fmt.Errorf("failed to get secured Connector client: mTLS HTTP client not initialized")
	}

	gqlClient, err := cp.gqlClientConstructor(cp.mtlsHTTPClient, cp.connectorSecuredURL, cp.enableLogging)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}
