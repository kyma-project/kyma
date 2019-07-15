package main

import (
	"fmt"
	kubelessClient "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	eventingClient "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"github.com/spf13/pflag"
	coreClient "k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testsuite"
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

func (s *scenario) SetGatewayHttpClient(httpClient *http.Client) {
	gatewayUrl := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, consts.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayUrl, httpClient)
	s.eventSender = testkit.NewEventSender(httpClient, s.Domain)
}

func (s *scenario) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

func (s *scenario) GetEventSender() *testkit.EventSender {
	return s.eventSender
}

func main() {
	time.Sleep(10 * time.Second)
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
	testNamespace := pflag.String("testNamespace", "default", "Namespace where test should create resources")
	domain := pflag.String("domain", "kyma.local", "Domain")
	cleanupOnly := pflag.Bool("cleanupOnly", false, "Only cleanup resources")
	skipCleanup := pflag.Bool("skipCleanup", false, "Do not cleanup resources")
	kubeconfigFlags := genericclioptions.NewConfigFlags()
	kubeconfigFlags.AddFlags(pflag.CommandLine)
	pflag.Parse()

	config, err := kubeconfigFlags.ToRESTConfig()
	if err != nil {
		log.Fatal(err)
	}

	runner := step.NewRunner()

	k8sResourceClient, err := resourceskit.NewK8sResourcesClient(config, *testNamespace)
	if err != nil {
		log.Fatal(err)
	}

	appOperatorClientset := appOperatorClient.NewForConfigOrDie(config)
	appBrokerClientset := appBrokerClient.NewForConfigOrDie(config)
	kubelessClientset := kubelessClient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	eventingClientset := eventingClient.NewForConfigOrDie(config)
	serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(config)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	tokenRequestClient, err := resourceskit.NewTokenRequestClient(config, *testNamespace)
	if err != nil {
		log.Fatal(err)
	}

	connector := testkit.NewConnectorClient(tokenRequestClient, true, log.New())

	s := &scenario{Domain: *domain}

	testService, err := testkit.NewTestService(k8sResourceClient, gatewayClientset.GatewayV1alpha2(), *domain)
	if err != nil {
		log.Fatal(err)
	}

	pods := coreClientset.CoreV1().Pods(*testNamespace)

	var steps = []step.Step{
		testsuite.NewCreateApplication(appOperatorClientset.ApplicationconnectorV1alpha1(), false),
		testsuite.NewCreateMapping(appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(*testNamespace)),
		testsuite.NewCreateEventActivation(appBrokerClientset.ApplicationconnectorV1alpha1().EventActivations(*testNamespace)),
		testsuite.NewDeployLambda(kubelessClientset.KubelessV1beta1().Functions(*testNamespace), pods),
		testsuite.NewStartTestServer(testService),
		testsuite.NewConnectApplication(connector, s),
		testsuite.NewRegisterTestService(testService, s),
		testsuite.NewCreateServiceInstance(serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(*testNamespace), s),
		testsuite.NewCreateServiceBinding(serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(*testNamespace), *testNamespace),
		testsuite.NewCreateServiceBindingUsage(serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(*testNamespace), pods, s),
		testsuite.NewCreateSubscription(eventingClientset.EventingV1alpha1().Subscriptions(*testNamespace), *testNamespace),
		testsuite.NewSendEvent(s),
		testsuite.NewCheckCounterPod(testService),
	}

	if *cleanupOnly {
		runner.Cleanup(steps)
		return
	}

	err = runner.Run(steps, *skipCleanup)

	if err != nil {
		os.Exit(1)
	}

	log.Info("Successfully Finished the e2e test!!")
}

func waitForApiSerer() {
	time.Sleep(10 * time.Second)
}

func setupLogging() {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
}

func setupFlags() {

}
