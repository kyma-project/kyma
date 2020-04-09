package compass

import (
	"k8s.io/client-go/rest"

	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"

	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sbuclientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
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

	kymaClients := testkit.InitKymaClients(config, s.testID)
	compassClients := testkit.InitCompassClients(kymaClients, state, s.domain, s.skipSSLVerify)
	knativeEventingClientset := eventingclientset.NewForConfigOrDie(config)
	serviceBindingUsageClientset := sbuclientset.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		kymaClients.CoreClientset.AppsV1().Deployments(s.testID),
		kymaClients.CoreClientset.CoreV1().Services(s.testID),
		kymaClients.ApiRules,
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.InClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, kymaClients.CoreClientset.CoreV1().Namespaces()),
		),
		step.Parallel(
			testsuite.NewStartTestServer(testService),
			testsuite.NewRegisterApplicationInCompass(s.testID,
				testService.GetInClusterTestServiceURL(),
				kymaClients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				compassClients.DirectorClient,
				httpSourceClientset.HTTPSources(helpers.KymaIntegrationNamespace),
				state),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, kymaClients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployFakeLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort,
				kymaClients.CoreClientset.AppsV1().Deployments(s.testID),
				kymaClients.CoreClientset.CoreV1().Services(s.testID),
				kymaClients.CoreClientset.CoreV1().Pods(s.testID),
				false),
			testsuite.NewConnectApplicationUsingCompass(compassClients.ConnectorClient, compassClients.DirectorClient, state),
		),
		testsuite.NewCreateServiceInstance(s.testID, s.testID, state.GetServiceClassID, state.GetServicePlanID,
			kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID),
			kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServicePlans(s.testID),
		),
		testsuite.NewCreateServiceBinding(s.testID, s.testID, kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientset.EventingV1alpha1().Brokers(s.testID), knativeEventingClientset.MessagingV1alpha1().Subscriptions(helpers.KymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, helpers.DefaultBrokerName, lambdaEndpoint, knativeEventingClientset.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSendEventToMesh(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 1),
		testsuite.NewSendEventToCompatibilityLayer(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 2),
	}, nil
}
