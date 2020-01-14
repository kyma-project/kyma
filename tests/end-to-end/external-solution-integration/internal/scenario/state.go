package scenario

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type e2eState struct {
	domain        string
	skipSSLVerify bool
	appName       string

	apiServiceInstanceName   string
	eventServiceInstanceName string
	eventSender              *testkit.EventSender
}

// SetAPIServiceInstanceName allows to set APIServiceInstanceName so it can be shared between steps
func (s *e2eState) SetAPIServiceInstanceName(serviceID string) {
	s.apiServiceInstanceName = serviceID
}

// SetEventServiceInstanceName allows to set EventServiceInstanceName so it can be shared between steps
func (s *e2eState) SetEventServiceInstanceName(serviceID string) {
	s.eventServiceInstanceName = serviceID
}

// GetAPIServiceInstanceName allows to get APIServiceInstanceName so it can be shared between steps
func (s *e2eState) GetAPIServiceInstanceName() string {
	return s.apiServiceInstanceName
}

// GetEventServiceInstanceName allows to get EventServiceInstanceName so it can be shared between steps
func (s *e2eState) GetEventServiceInstanceName() string {
	return s.eventServiceInstanceName
}

// GetEventSender returns connected EventSender
func (s *e2eState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}

type appConnectorE2EState struct {
	e2eState

	serviceClassID string
	registryClient *testkit.RegistryClient
}

func (s *E2E) NewState() *appConnectorE2EState {
	return &appConnectorE2EState{e2eState: e2eState{domain: s.domain, skipSSLVerify: s.skipSSLVerify, appName: s.testID}}
}

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *appConnectorE2EState) SetServiceClassID(serviceID string) {
	s.serviceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *appConnectorE2EState) GetServiceClassID() string {
	return s.serviceClassID
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *appConnectorE2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.skipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.domain, s.appName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHTTPClient)
	s.eventSender = testkit.NewEventSender(resilientHTTPClient, s.domain)
}

// GetRegistryClient returns connected RegistryClient
func (s *appConnectorE2EState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

type compassE2EState struct {
	e2eState

	compassAppID string
	config       compassEnvConfig
}

type compassEnvConfig struct {
	Tenant             string
	ScenariosLabelKey  string `envconfig:"default=scenarios"`
	DexSecretName      string
	DexSecretNamespace string
	RuntimeID          string
}

func (s *CompassE2E) NewState() (*compassE2EState, error) {
	config := compassEnvConfig{}
	err := envconfig.Init(&config)
	if err != nil {
		return nil, errors.Wrap(err, "while loading environment variables")
	}

	return &compassE2EState{
		e2eState: e2eState{domain: s.domain, skipSSLVerify: s.skipSSLVerify, appName: s.testID},
		config:   config,
	}, nil
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *compassE2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.skipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	s.eventSender = testkit.NewEventSender(resilientHTTPClient, s.domain)
}

// GetCompassAppID returns Compass ID of registered application
func (s *compassE2EState) GetCompassAppID() string {
	return s.compassAppID
}

// SetCompassAppID sets Compass ID of registered application
func (s *compassE2EState) SetCompassAppID(appID string) {
	s.compassAppID = appID
}

// GetScenariosLabelKey returns Compass label key for scenarios label
func (s *compassE2EState) GetScenariosLabelKey() string {
	return s.config.ScenariosLabelKey
}

// GetDefaultTenant returns Compass ID of tenant that is used for tests
func (s *compassE2EState) GetDefaultTenant() string {
	return s.config.Tenant
}

// GetRuntimeID returns Compass ID of runtime that is tested
func (s *compassE2EState) GetRuntimeID() string {
	return s.config.RuntimeID
}

// GetDexSecret returns name and namespace of secret with dex account
func (s *compassE2EState) GetDexSecret() (string, string) {
	return s.config.DexSecretName, s.config.DexSecretNamespace
}
