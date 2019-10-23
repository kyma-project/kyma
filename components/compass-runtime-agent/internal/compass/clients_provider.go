package compass

import (
	"github.com/pkg/errors"
	"kyma-project.io/compass-runtime-agent/internal/certificates"
	"kyma-project.io/compass-runtime-agent/internal/compass/connector"
	"kyma-project.io/compass-runtime-agent/internal/compass/director"
	"kyma-project.io/compass-runtime-agent/internal/config"
	"kyma-project.io/compass-runtime-agent/internal/graphql"
)

//go:generate mockery -name=ClientsProvider
type ClientsProvider interface {
	GetDirectorClient(credentials certificates.ClientCredentials, url string, runtimeConfig config.RuntimeConfig) (director.DirectorClient, error)
	GetConnectorClient(url string) (connector.Client, error)
	GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.Client, error)
}

func NewClientsProvider(gqlClientConstr graphql.ClientConstructor, insecureConnectorCommunication, insecureConfigFetch, enableLogging bool) ClientsProvider {
	return &clientsProvider{
		gqlClientConstructor:            gqlClientConstr,
		insecureConnectionCommunication: insecureConnectorCommunication,
		insecureConfigFetch:             insecureConfigFetch,
		enableLogging:                   enableLogging,
	}
}

type clientsProvider struct {
	gqlClientConstructor            graphql.ClientConstructor
	insecureConnectionCommunication bool
	insecureConfigFetch             bool
	enableLogging                   bool
}

func (cp *clientsProvider) GetDirectorClient(credentials certificates.ClientCredentials, url string, runtimeConfig config.RuntimeConfig) (director.DirectorClient, error) {
	gqlClient, err := cp.gqlClientConstructor(credentials.AsTLSCertificate(), url, cp.enableLogging, cp.insecureConfigFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return director.NewConfigurationClient(gqlClient, runtimeConfig), nil
}

func (cp *clientsProvider) GetConnectorClient(url string) (connector.Client, error) {
	gqlClient, err := cp.gqlClientConstructor(nil, url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}

func (cp *clientsProvider) GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.Client, error) {
	gqlClient, err := cp.gqlClientConstructor(credentials.AsTLSCertificate(), url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}
