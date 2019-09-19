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
	GetConnectorTokenSecuredClient(url string) (connector.TokenSecuredClient, error)
	GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.CertificateSecuredClient, error)
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

func (cp *clientsProvider) GetConnectorTokenSecuredClient(url string) (connector.TokenSecuredClient, error) {
	gqlClient, err := cp.gqlClientConstructor(nil, url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewTokenSecuredConnectorClient(gqlClient), nil
}

func (cp *clientsProvider) GetConnectorCertSecuredClient(credentials certificates.ClientCredentials, url string) (connector.CertificateSecuredClient, error) {
	gqlClient, err := cp.gqlClientConstructor(credentials.AsTLSCertificate(), url, cp.enableLogging, cp.insecureConnectionCommunication)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create GraphQL client")
	}

	return connector.NewCertificateSecuredConnectorClient(gqlClient), nil
}
