package scenario

import (
	"crypto/tls"
	"fmt"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type E2EState struct {
	Domain           string
	GatewaySubdomain string
	SkipSSLVerify    bool
	AppName          string

	ServiceClassID           string
	ApiServiceInstanceName   string
	EventServiceInstanceName string
	EventSender              *testkit.EventSender
	RegistryClient           *testkit.RegistryClient
}

// GetRegistryClient returns connected RegistryClient
func (s *E2EState) GetRegistryClient() *testkit.RegistryClient {
	return s.RegistryClient
}

// SetAPIServiceInstanceName allows to set APIServiceInstanceName so it can be shared between steps
func (s *E2EState) SetAPIServiceInstanceName(serviceID string) {
	s.ApiServiceInstanceName = serviceID
}

// SetEventServiceInstanceName allows to set EventServiceInstanceName so it can be shared between steps
func (s *E2EState) SetEventServiceInstanceName(serviceID string) {
	s.EventServiceInstanceName = serviceID
}

// GetAPIServiceInstanceName allows to get APIServiceInstanceName so it can be shared between steps
func (s *E2EState) GetAPIServiceInstanceName() string {
	return s.ApiServiceInstanceName
}

// GetEventServiceInstanceName allows to get EventServiceInstanceName so it can be shared between steps
func (s *E2EState) GetEventServiceInstanceName() string {
	return s.EventServiceInstanceName
}

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *E2EState) SetServiceClassID(serviceID string) {
	s.ServiceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *E2EState) GetServiceClassID() string {
	return s.ServiceClassID
}

// GetEventSender returns connected EventSender
func (s *E2EState) GetEventSender() *testkit.EventSender {
	return s.EventSender
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *E2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	gatewayURL := fmt.Sprintf("https://%s.%s/%s/v1/metadata/services", s.GatewaySubdomain, s.Domain, s.AppName)
	httpClient := internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify), internal.WithClientCertificates(certs))
	s.RegistryClient = testkit.NewRegistryClient(gatewayURL, resilient.WrapHttpClient(httpClient))
	s.EventSender = testkit.NewEventSender(httpClient, s.Domain, s.AppName)
}
