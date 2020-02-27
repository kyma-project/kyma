package e2e

import (
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

// Steps return scenario steps
func (s *E2EScenario) Steps(config *rest.Config) ([]step.Step, error) {
	clients := testkit.InitKymaClients(config, s.testID)

	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	appConnector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(s.SkipSSLVerify),
		log.New(),
	)

	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.SkipSSLVerify),
		clients.CoreClientset.AppsV1().Deployments(s.testID),
		clients.CoreClientset.CoreV1().Services(s.testID),
		clients.GatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.Domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	state := s.NewState()

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, clients.CoreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.testID, s.testID, false, s.ApplicationTenant, s.ApplicationGroup,
				clients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(helpers.KymaIntegrationNamespace))),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, clients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort, clients.KubelessClientset.KubelessV1beta1().Functions(s.testID), clients.Pods),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(appConnector, state, s.ApplicationTenant, s.ApplicationGroup),
		),
		testsuite.NewRegisterTestService(s.testID, testService, state),
		testsuite.NewCreateServiceInstance(s.testID, s.testID, state.GetServiceClassID,
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID),
		),
		testsuite.NewCreateServiceBinding(s.testID, s.testID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			clients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID),
			knativeEventingClientSet.MessagingV1alpha1().Subscriptions(helpers.KymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, helpers.DefaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSendEvent(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
