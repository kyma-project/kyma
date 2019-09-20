package compass

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/connector"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/pkg/errors"
)

type ClientsProvider interface {
	GetCompassConfigClient(credentials certificates.ClientCredentials, url string) (director.ConfigClient, error)
	GetConnectorClient(url string) (connector.ConnectorClient, error)
	GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.ConnectorClient, error)
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

func (cp *clientsProvider) GetCompassConfigClient(credentials certificates.ClientCredentials, url string) (director.ConfigClient, error) {
	gqlClient, err := cp.gqlClientConstructor(credentials.AsTLSCertificate(), url, cp.enableLogging, cp.insecureConfigFetch)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return director.NewConfigurationClient(gqlClient), nil
}

func (cp *clientsProvider) GetConnectorClient(url string) (connector.ConnectorClient, error) {
	gqlClient, err := cp.gqlClientConstructor(nil, url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}

func (cp *clientsProvider) GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.ConnectorClient, error) {
	gqlClient, err := cp.gqlClientConstructor(credentials.AsTLSCertificate(), url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewConnectorClient(gqlClient), nil
}
