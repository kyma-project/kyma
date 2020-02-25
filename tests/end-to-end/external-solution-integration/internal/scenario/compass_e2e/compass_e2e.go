package compass_e2e

import (
	"fmt"

	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"k8s.io/client-go/rest"
)

// Steps return scenario steps
func (s *CompassE2EScenario) Steps(config *rest.Config) ([]step.Step, error) {
	state, err := s.NewState()
	if err != nil {
		return nil, err
	}

	kymaClients := testkit.InitKymaClients(config, s.testID)
	compassClients := testkit.InitCompassClients(kymaClients, state, s.domain, s.skipSSLVerify)
	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		kymaClients.CoreClientset.AppsV1().Deployments(s.testID),
		kymaClients.CoreClientset.CoreV1().Services(s.testID),
		kymaClients.GatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	apiServiceInstanceName := fmt.Sprintf("%s-api", s.testID)

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
				state),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, kymaClients.AppBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort, kymaClients.KubelessClientset.KubelessV1beta1().Functions(s.testID), kymaClients.Pods),
			testsuite.NewConnectApplicationUsingCompass(compassClients.ConnectorClient, compassClients.DirectorClient, state),
		),
		testsuite.NewCreateSeparateServiceInstance(s.testID, kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			kymaClients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(), state,
		),
		testsuite.NewCreateServiceBinding(s.testID, apiServiceInstanceName, kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(helpers.KymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, helpers.DefaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSendEventToMesh(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
