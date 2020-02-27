package e2e

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type e2EState struct {
	scenario.E2EState

	serviceClassID string
	registryClient *testkit.RegistryClient
}

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *e2EState) SetServiceClassID(serviceID string) {
	s.serviceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *e2EState) GetServiceClassID() string {
	return s.serviceClassID
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.SkipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, s.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHTTPClient)
	s.EventSender = testkit.NewEventSender(resilientHTTPClient, s.Domain, nil)
}

// GetRegistryClient returns connected RegistryClient
func (s *e2EState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}
