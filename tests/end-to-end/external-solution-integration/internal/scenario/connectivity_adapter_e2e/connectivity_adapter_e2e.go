package connectivity_adapter_e2e

import (
	"fmt"

	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"

	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

// Steps return scenario steps
func (s *CompassConnectivityAdapterE2EConfig) Steps(config *rest.Config) ([]step.Step, error) {
	state, err := s.NewState()
	if err != nil {
		return nil, err
	}

	clients := testkit.InitKymaClients(config, s.testID)
	compassClients := testkit.InitCompassClients(clients, state, s.domain, s.skipSSLVerify)

	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)

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

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	apiServiceInstanceName := fmt.Sprintf("%s-api", s.testID)
	eventServiceInstanceName := fmt.Sprintf("%s-event", s.testID)

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, clients.CoreClientset.CoreV1().Namespaces()),
			testsuite.RegisterEmptyApplicationInCompass(s.testID, compassClients.DirectorClient, state),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, clients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewStartTestServer(testService),
			testsuite.NewRegisterServiceUsingConnectivityAdapter(compassClients.ConnectorClient, appConnector, compassClients.DirectorClient, state),
			testsuite.NewDeployLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort, clients.KubelessClientset.KubelessV1beta1().Functions(s.testID), clients.Pods),
		),
		testsuite.NewRegisterLegacyServiceInCompass(s.testID, testService.GetInClusterTestServiceURL(), compassClients.DirectorClient, testService, state),
		step.Parallel(
			testsuite.NewCreateServiceInstance(s.testID, apiServiceInstanceName, state.GetApiServiceClassID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
				clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
			testsuite.NewCreateServiceInstance(s.testID, eventServiceInstanceName, state.GetEventServiceClassID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
				clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
		),
		testsuite.NewCreateServiceBinding(s.testID, apiServiceInstanceName, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			clients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID),
			knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, defaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSendEvent(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
