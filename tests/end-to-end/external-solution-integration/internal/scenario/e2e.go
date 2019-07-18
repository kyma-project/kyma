package scenario

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/common/resilient"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// E2E executes complete external solution integration test scenario
type E2E struct {
	domain        string
	runID         string
	testNamespace string
	skipSSLVerify bool
}

type e2EState struct {
	domain        string
	skipSSLVerify bool

	serviceClassID      string
	serviceInstanceName string
	registryClient      *testkit.RegistryClient
	eventSender         *testkit.EventSender
}

// AddFlags adds CLI flags to given FlagSet
func (s *E2E) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.testNamespace, "testNamespace", "default", "Namespace where test should create resources")
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.runID, "runID", "e2e-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skip-ssl-verify", false, "Skip verification of service SSL certificates")
}

// Steps return scenario steps
func (s *E2E) Steps(config *rest.Config) ([]step.Step, error) {
	appOperatorClientset := appOperatorClient.NewForConfigOrDie(config)
	appBrokerClientset := appBrokerClient.NewForConfigOrDie(config)
	kubelessClientset := kubelessClient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	pods := coreClientset.CoreV1().Pods(s.testNamespace)
	eventingClientset := eventingClient.NewForConfigOrDie(config)
	serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(config)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	connector := testkit.NewConnectorClient(
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testNamespace),
		internal.NewHTTPClient(s.skipSSLVerify),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		coreClientset.AppsV1().Deployments(s.testNamespace),
		coreClientset.CoreV1().Services(s.testNamespace),
		gatewayClientset.GatewayV1alpha2().Apis(s.testNamespace),
		s.domain,
		s.testNamespace,
	)

	state := &e2EState{
		domain:        s.domain,
		skipSSLVerify: s.skipSSLVerify,
	}

	return []step.Step{
		testsuite.NewCreateNamespace(coreClientset.CoreV1().Namespaces(), s.testNamespace),
		testsuite.NewCreateApplication(appOperatorClientset.ApplicationconnectorV1alpha1().Applications(), false),
		testsuite.NewCreateMapping(appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testNamespace)),
		testsuite.NewDeployLambda(kubelessClientset.KubelessV1beta1().Functions(s.testNamespace), pods),
		testsuite.NewStartTestServer(testService),
		testsuite.NewConnectApplication(connector, state),
		testsuite.NewRegisterTestService(testService, state),
		testsuite.NewCreateServiceInstance(s.runID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testNamespace), state),
		testsuite.NewCreateServiceBinding(serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testNamespace), state),
		testsuite.NewCreateServiceBindingUsage(serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testNamespace), pods, state),
		testsuite.NewCreateSubscription(eventingClientset.EventingV1alpha1().Subscriptions(s.testNamespace), s.testNamespace),
		testsuite.NewSleep(20 * time.Second),
		testsuite.NewSendEvent(state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *e2EState) SetServiceClassID(serviceID string) {
	s.serviceClassID = serviceID
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *e2EState) GetServiceClassID() string {
	return s.serviceClassID
}

// SetServiceInstanceName allows to set ServiceInstanceName so it can be shared between steps
func (s *e2EState) SetServiceInstanceName(serviceID string) {
	s.serviceInstanceName = serviceID
}

// GetServiceInstanceName allows to get ServiceInstanceName so it can be shared between steps
func (s *e2EState) GetServiceInstanceName() string {
	return s.serviceInstanceName
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EState) SetGatewayClientCerts(certs []tls.Certificate) {
	httpClient := internal.NewHTTPClient(s.skipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.domain, consts.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHTTPClient)
	s.eventSender = testkit.NewEventSender(resilientHTTPClient, s.domain)
}

// GetRegistryClient returns connected RegistryClient
func (s *e2EState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

// GetEventSender returns connected EventSender
func (s *e2EState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}
