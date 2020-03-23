package send_and_check_event_only

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type state struct {
	scenario.E2EState
	registryClient *testkit.RegistryClient
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *state) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.SkipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, s.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHTTPClient)
	s.EventSender = testkit.NewEventSender(resilientHTTPClient, s.Domain, nil)
}

// GetRegistryClient returns connected RegistryClient
func (s *state) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}
