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
	clients := testkit.InitKymaClients(config, s.TestID)

	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	appConnector := testkit.NewConnectorClient(
		s.TestID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.TestID),
		internal.NewHTTPClient(s.SkipSSLVerify),
		log.New(),
	)

	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.SkipSSLVerify),
		clients.CoreClientset.AppsV1().Deployments(s.TestID),
		clients.CoreClientset.CoreV1().Services(s.TestID),
		clients.GatewayClientset.GatewayV1alpha2().Apis(s.TestID),
		s.Domain,
		s.TestID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.TestID, s.TestID, helpers.LambdaPort)
	state := s.NewState()

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.TestID, clients.CoreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.TestID, s.TestID, false, s.ApplicationTenant, s.ApplicationGroup,
				clients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(helpers.KymaIntegrationNamespace))),
		step.Parallel(
			testsuite.NewCreateMapping(s.TestID, clients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.TestID)),
			testsuite.NewDeployLambda(s.TestID, helpers.LambdaPayload, helpers.LambdaPort, clients.KubelessClientset.KubelessV1beta1().Functions(s.TestID), clients.Pods),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(appConnector, state, s.ApplicationTenant, s.ApplicationGroup),
		),
		testsuite.NewRegisterTestService(s.TestID, testService, state),
		testsuite.NewCreateServiceInstance(s.TestID, s.TestID, state.GetServiceClassID,
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.TestID),
			clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.TestID),
		),
		testsuite.NewCreateServiceBinding(s.TestID, s.TestID, clients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.TestID)),
		testsuite.NewCreateServiceBindingUsage(s.TestID, s.TestID, s.TestID,
			clients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.TestID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.TestID),
			knativeEventingClientSet.MessagingV1alpha1().Subscriptions(helpers.KymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.TestID, helpers.DefaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.TestID)),
		testsuite.NewSendEvent(s.TestID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
