package scenario

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

// E2E executes complete external solution integration test scenario
type E2E struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
}

const (
	lambdaPort = 8080
)

// AddFlags adds CLI flags to given FlagSet
func (s *E2E) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "e2e-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
}

// Steps return scenario steps
func (s *E2E) Steps(config *rest.Config) ([]step.Step, error) {
	clients := testkit.InitClients(config, s.testID)

	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	appConnector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(s.skipSSLVerify),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		clients.CoreClientset.AppsV1().Deployments(s.testID),
		clients.CoreClientset.CoreV1().Services(s.testID),
		clients.GatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, lambdaPort)
	state := s.NewState()

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, clients.CoreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant, s.applicationGroup, clients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications()),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, clients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, lambdaPort, clients.KubelessClientset.KubelessV1beta1().Functions(s.testID), clients.Pods),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(appConnector, state, s.applicationTenant, s.applicationGroup),
		),
		testsuite.NewRegisterTestService(s.testID, testService, state),
		testsuite.NewCreateServiceInstance(s.testID,
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID),
			state,
		),
		testsuite.NewCreateServiceBinding(s.testID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID), state),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID, clients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID)),
		testsuite.NewCreateSubscription(s.testID, s.testID, lambdaEndpoint, clients.EventingClientset.EventingV1alpha1().Subscriptions(s.testID)),
		testsuite.NewSendEvent(s.testID, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
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
