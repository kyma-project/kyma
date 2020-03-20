package compass_e2e

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type compassE2EState struct {
	scenario.E2EState
	scenario.CompassEnvConfig
	compassAppID         string
	connectivityURL      string
	servicePlanID        string
	registryClient       *testkit.RegistryClient
	legacyRegistryClient *testkit.LegacyRegistryClient
	cert                 []tls.Certificate
}

// GetCompassAppID returns Compass ID of registered application
func (s *compassE2EState) GetCompassAppID() string {
	return s.compassAppID
}

// SetCompassAppID sets Compass ID of registered application
func (s *compassE2EState) SetCompassAppID(appID string) {
	s.compassAppID = appID
}

func (s *compassE2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	metadataURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, s.AppName)
	httpClient := internal.NewHTTPClient(s.SkipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	s.registryClient = testkit.NewRegistryClient(metadataURL, resilientHTTPClient)
	s.EventSender = testkit.NewEventSender(resilientHTTPClient, s.Domain, nil)
}

func (s *compassE2EState) GetServicePlanID() string {
	return s.servicePlanID
}

func (s *compassE2EState) SetServicePlanID(servicePlanID string) {
	s.servicePlanID = servicePlanID
}
