package event_mesh_2_phases_test

import (
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	coreClient "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

const (
	kymaIntegrationNamespace = "kyma-integration"
	defaultBrokerName        = "default"
)

// Steps return scenario steps
func (s *TwoPhasesEventMeshTestConfig) Steps(config *rest.Config) ([]step.Step, error) {
	//appOperatorClientset := appOperatorClient.NewForConfigOrDie(config)
	//appBrokerClientset := appBrokerClient.NewForConfigOrDie(config)
	//kubelessClientset := kubelessClient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	//pods := coreClientset.CoreV1().Pods(s.testID)
	//serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(config)
	//serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	//connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	//httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	//knativeEventingClientSet := eventing.NewForConfigOrDie(config)

	//connector := testkit.NewConnectorClient(
	//	s.testID,
	//	connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
	//	internal.NewHTTPClient(s.skipSSLVerify),
	//	log.New(),
	//)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		gatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	//lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	state := s.NewState()
	dataStore := testkit.NewDataStore(coreClientset, s.testID)
	state.SetDataStore(dataStore)

	return []step.Step{
		//step.Parallel(
		//	testsuite.NewCreateNamespace(s.testID, coreClientset.CoreV1().Namespaces()),
		//	testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant,
		//		s.applicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
		//		httpSourceClientset.HTTPSources(kymaIntegrationNamespace)),
		//),
		//step.Parallel(
		//	testsuite.NewCreateMapping(s.testID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
		//	testsuite.NewDeployLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort, kubelessClientset.KubelessV1beta1().Functions(s.testID), pods),
		//	testsuite.NewStartTestServer(testService),
		//	testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
		//),
		//testsuite.NewRegisterTestService(s.testID, testService, state),
		//testsuite.NewCreateServiceInstance(s.testID, s.testID, state.GetServiceClassID,
		//	serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
		//	serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
		//testsuite.NewCreateServiceBinding(s.testID, s.testID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		//testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
		//	serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
		//	knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
		//testsuite.NewCreateKnativeTrigger(s.testID, defaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewReuseApplication(state),
		testsuite.NewSendEventToMesh(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
