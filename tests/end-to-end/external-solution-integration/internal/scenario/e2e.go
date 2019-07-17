package scenario

import (
	"fmt"
	"github.com/kyma-project/kyma/common/resilient"
	"time"

	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/common/ingressgateway"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
)

type E2E struct {
	domain        string
	runID         string
	testNamespace string
}
type e2EState struct {
	domain string

	serviceClassID      string
	serviceInstanceName string
	registryClient      *testkit.RegistryClient
	eventSender         *testkit.EventSender
}

func (s *E2E) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.testNamespace, "testNamespace", "default", "Namespace where test should create resources")
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.runID, "runID", "e2e-test", "domain")
}

func (s *E2E) Steps(config *rest.Config) ([]step.Step, error) {
	k8sResourceClient, err := resourceskit.NewK8sResourcesClient(config, s.testNamespace)
	if err != nil {
		log.Fatal(err)
	}

	ingressHTTPClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		log.Fatal(err)
	}

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
	tokenRequestClient := resourceskit.NewTokenRequestClient(connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testNamespace))
	connector := testkit.NewConnectorClient(tokenRequestClient, true, log.New())
	testService := testkit.NewTestService(k8sResourceClient, ingressHTTPClient, gatewayClientset.GatewayV1alpha2(), s.domain, s.testNamespace)

	state := &e2EState{domain: s.domain}

	return []step.Step{
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

func (s *e2EState) SetServiceClassID(serviceID string) {
	s.serviceClassID = serviceID
}

func (s *e2EState) GetServiceClassID() string {
	return s.serviceClassID
}

func (s *e2EState) SetServiceInstanceName(serviceID string) {
	s.serviceInstanceName = serviceID
}

func (s *e2EState) GetServiceInstanceName() string {
	return s.serviceInstanceName
}

func (s *e2EState) SetGatewayHTTPClient(httpClient *http.Client) {
	resilientHttpClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.domain, consts.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHttpClient)
	s.eventSender = testkit.NewEventSender(resilientHttpClient, s.domain)
}

func (s *e2EState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

func (s *e2EState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}
