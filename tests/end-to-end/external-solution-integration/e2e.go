package main

import (
	"fmt"

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
	"github.com/spf13/pflag"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"os"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type scenario struct {
	Domain         string
	serviceID      string
	registryClient *testkit.RegistryClient
	eventSender    *testkit.EventSender
}

func (s *scenario) SetServiceID(serviceID string) {
	s.serviceID = serviceID
}

func (s *scenario) GetServiceID() string {
	return s.serviceID
}

func (s *scenario) SetGatewayHTTPClient(httpClient *http.Client) {
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, consts.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, httpClient)
	s.eventSender = testkit.NewEventSender(httpClient, s.Domain)
}

func (s *scenario) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

func (s *scenario) GetEventSender() *testkit.EventSender {
	return s.eventSender
}

var (
	testNamespace string
	domain        string
	kubeConfig    *rest.Config
	runner        *step.Runner
)

func main() {
	runner = step.NewRunner()
	waitForAPIServer()
	setupLogging()
	setupFlags()

	k8sResourceClient, err := resourceskit.NewK8sResourcesClient(kubeConfig, testNamespace)
	if err != nil {
		log.Fatal(err)
	}

	ingressHTTPClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		log.Fatal(err)
	}

	appOperatorClientset := appOperatorClient.NewForConfigOrDie(kubeConfig)
	appBrokerClientset := appBrokerClient.NewForConfigOrDie(kubeConfig)
	kubelessClientset := kubelessClient.NewForConfigOrDie(kubeConfig)
	coreClientset := coreClient.NewForConfigOrDie(kubeConfig)
	pods := coreClientset.CoreV1().Pods(testNamespace)
	eventingClientset := eventingClient.NewForConfigOrDie(kubeConfig)
	serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(kubeConfig)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(kubeConfig)
	gatewayClientset := gatewayClient.NewForConfigOrDie(kubeConfig)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(kubeConfig)
	tokenRequestClient := resourceskit.NewTokenRequestClient(connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(testNamespace))
	connector := testkit.NewConnectorClient(tokenRequestClient, true, log.New())
	testService := testkit.NewTestService(k8sResourceClient, ingressHTTPClient, gatewayClientset.GatewayV1alpha2(), domain)

	s := &scenario{Domain: domain}

	var steps = []step.Step{
		testsuite.NewCreateApplication(appOperatorClientset.ApplicationconnectorV1alpha1(), false),
		testsuite.NewCreateMapping(appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(testNamespace)),
		testsuite.NewDeployLambda(kubelessClientset.KubelessV1beta1().Functions(testNamespace), pods),
		testsuite.NewStartTestServer(testService),
		testsuite.NewConnectApplication(connector, s),
		testsuite.NewRegisterTestService(testService, s),
		testsuite.NewCreateServiceInstance(serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(testNamespace), s),
		testsuite.NewCreateServiceBinding(serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(testNamespace), testNamespace),
		testsuite.NewCreateServiceBindingUsage(serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(testNamespace), pods, s),
		testsuite.NewCreateSubscription(eventingClientset.EventingV1alpha1().Subscriptions(testNamespace), testNamespace),
		testsuite.NewSendEvent(s),
		testsuite.NewCheckCounterPod(testService),
	}

	err = runner.Execute(steps)

	if err != nil {
		os.Exit(1)
	}

	log.Info("Successfully Finished the e2e test!!")
}

func waitForAPIServer() {
	time.Sleep(10 * time.Second)
}

func setupLogging() {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
}

func setupFlags() {
	var err error
	pflag.StringVar(&testNamespace, "testNamespace", "default", "Namespace where test should create resources")
	pflag.StringVar(&domain, "domain", "kyma.local", "Domain")
	kubeconfigFlags := genericclioptions.NewConfigFlags()
	kubeconfigFlags.AddFlags(pflag.CommandLine)
	runner.AddFlags(pflag.CommandLine)
	pflag.Parse()

	kubeConfig, err = kubeconfigFlags.ToRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
}
