package scenario

import (
	"crypto/tls"
	"net/http"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

// CompassE2E executes complete external solution integration test scenario
// using Compass for Application registration and connectivity
type CompassE2E struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
	lambdaPort        int
}

// AddFlags adds CLI flags to given FlagSet
func (s *CompassE2E) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "compass-e2e-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
	pflag.IntVar(&s.lambdaPort, "lambdaPort", 8080, "Lambda port")
}

// Steps return scenario steps
func (s *CompassE2E) Steps(config *rest.Config) ([]step.Step, error) {
	state, err := s.NewState()
	if err != nil {
		return nil, err
	}

	kymaClients := testkit.InitClients(config, s.testID)
	compassConnector := testkit.NewCompassConnectorClient(s.skipSSLVerify)
	compassDirector, err := testkit.NewCompassDirectorClientOrDie(kymaClients.CoreClientset, state, s.domain)
	if err != nil {
		return nil, err
	}

	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		kymaClients.CoreClientset.AppsV1().Deployments(s.testID),
		kymaClients.CoreClientset.CoreV1().Services(s.testID),
		kymaClients.GatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, s.lambdaPort)

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, kymaClients.CoreClientset.CoreV1().Namespaces()),
			testsuite.NewAssignScenarioInCompass(s.testID, state.GetRuntimeID(), compassDirector),
		),
		step.Parallel(
			testsuite.NewStartTestServer(testService),
			testsuite.NewRegisterApplicationInCompass(s.testID,
				testService.GetInClusterTestServiceURL(),
				kymaClients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				compassDirector,
				state),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, kymaClients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, s.lambdaPort, kymaClients.KubelessClientset.KubelessV1beta1().Functions(s.testID), kymaClients.Pods),
			testsuite.NewConnectApplicationUsingCompass(compassConnector, compassDirector, state),
		),
		testsuite.NewCreateSeparateServiceInstance(s.testID,
			kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			kymaClients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
			state,
		),
		testsuite.NewCreateServiceBinding(s.testID, kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID), state),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID, kymaClients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID)),
		testsuite.NewCreateSubscription(s.testID, s.testID, lambdaEndpoint, kymaClients.EventingClientset.EventingV1alpha1().Subscriptions(s.testID)),
		testsuite.NewSendEvent(s.testID, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
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
