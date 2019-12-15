package testkit

import (
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/clientset"
)


type CompassConnectorClient struct {
	connector *clientset.ConnectorClientSet
}

func NewCompassConnectorClient(skipTLSVerify bool) *CompassConnectorClient {
	return &CompassConnectorClient{
		connector: clientset.NewConnectorClientSet(clientset.WithSkipTLSVerify(skipTLSVerify)),
	}
}

func (cc *CompassConnectorClient) GenerateCertificateForToken(token, connectorURL string) (tls.Certificate, error) {
	return cc.connector.GenerateCertificateForToken(token, connectorURL)
}
