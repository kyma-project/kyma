package connectivity_adapter

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type state struct {
	scenario.E2EState
	scenario.CompassEnvConfig
	compassAppID         string
	servicePlanID        string
	registryClient       *testkit.RegistryClient
	legacyRegistryClient *testkit.LegacyRegistryClient
	cert                 []tls.Certificate
}

func (s *state) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

func (s *state) GetEventSender() *testkit.EventSender {
	return s.EventSender
}

func (s *state) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.SkipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	legacyMetadataURL := fmt.Sprintf("https://adapter-gateway-mtls.%s/%s/v1/metadata/services", s.Domain, s.AppName)
	s.registryClient = testkit.NewRegistryClient(legacyMetadataURL, resilientHTTPClient)
	s.EventSender = testkit.NewEventSender(resilientHTTPClient, s.Domain, nil)
}

// SetCompassAppID sets Compass ID of registered application
func (s *state) SetCompassAppID(appID string) {
	s.compassAppID = appID
}

// GetCompassAppID returns Compass ID of registered application
func (s *state) GetCompassAppID() string {
	return s.compassAppID
}

func (s *state) GetServicePlanID() string {
	return s.servicePlanID
}

func (s *state) SetServicePlanID(servicePlanID string) {
	s.servicePlanID = servicePlanID
}
