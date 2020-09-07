package connectivity_adapter

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"

	connectiontokenhandlerclientset "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	state, err := s.NewState()
	if err != nil {
		return nil, err
	}

	clients := testkit.InitKymaClients(config, s.testID)
	compassClients := testkit.InitCompassClients(clients, state, s.domain, s.skipSSLVerify)

	connectionTokenHandlerClientset := connectiontokenhandlerclientset.NewForConfigOrDie(config)
	knativeEventingClientSet := eventingclientset.NewForConfigOrDie(config)

	appConnector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		clients.CoreClientset.AppsV1().Deployments(s.testID),
		clients.CoreClientset.CoreV1().Services(s.testID),
		clients.ApiRules,
		s.domain,
		s.testID,
		s.testServiceImage,
	)

	functionEndpoint := helpers.InClusterEndpoint(s.testID, s.testID, helpers.FunctionPort)
	const eventId = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, clients.CoreClientset.CoreV1().Namespaces()),
			testsuite.RegisterEmptyApplicationInCompass(s.testID, compassClients.DirectorClient, state),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, clients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplicationUsingCompassLegacy(compassClients.ConnectorClient, appConnector, compassClients.DirectorClient, state),
			testsuite.NewDeployFunction(s.testID, helpers.FunctionPayload, helpers.FunctionPort, clients.Function, false),
		),
		testsuite.NewRegisterLegacyServiceInCompass(s.testID, testService.GetInClusterTestServiceURL(), testService, state),
		testsuite.NewCreateServiceInstance(s.testID, s.testID, state.GetServiceClassID, state.GetServicePlanID,
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID),
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServicePlans(s.testID),
		),
		testsuite.NewCreateServiceBinding(s.testID, s.testID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			clients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID),
			knativeEventingClientSet.MessagingV1alpha1().Subscriptions(helpers.KymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, helpers.DefaultBrokerName, functionEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSleep(s.eventSendDelay),
		testsuite.NewSendEventToMesh(s.testID, helpers.FunctionPayload, state, eventId),
		testsuite.NewCheckCounterPod(testService, 1),
		testsuite.NewSendEventToCompatibilityLayer(s.testID, helpers.FunctionPayload, state, eventId),
		testsuite.NewCheckCounterPod(testService, 2),
	}, nil
}
