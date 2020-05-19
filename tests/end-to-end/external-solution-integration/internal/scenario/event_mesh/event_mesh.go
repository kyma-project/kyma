package event_mesh

import (
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_evaluate"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	s.prepare = event_mesh_prepare.Scenario{
		Domain:            s.domain,
		TestID:            s.testID,
		SkipSSLVerify:     s.skipSSLVerify,
		ApplicationTenant: s.applicationTenant,
		ApplicationGroup:  s.applicationGroup,
	}
	s.evaluate = event_mesh_evaluate.Scenario{
		Domain:        s.domain,
		TestID:        s.testID,
		SkipSSLVerify: s.skipSSLVerify,
	}

	prepareSteps, err := s.prepare.Steps(config)
	if err != nil {
		return nil, err
	}
	evalSteps, err := s.evaluate.Steps(config)
	if err != nil {
		return nil, err
	}
	return append(prepareSteps, evalSteps...), nil
	//	if level, err := log.ParseLevel(s.logLevel); err != nil {
	//		return nil, fmt.Errorf("False \"logLevel\" configuration: %v", err)
	//	} else {
	//		log.SetLevel(level)
	//	}
	//	appOperatorClientset := appoperatorclientset.NewForConfigOrDie(config)
	//	appBrokerClientset := appbrokerclientset.NewForConfigOrDie(config)
	//	coreClientset := k8s.NewForConfigOrDie(config)
	//	serviceCatalogClientset := servicecatalogclientset.NewForConfigOrDie(config)
	//	serviceBindingUsageClientset := sbuclientset.NewForConfigOrDie(config)
	//	connectionTokenHandlerClientset := connectiontokenhandlerclientset.NewForConfigOrDie(config)
	//	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	//	knativeEventingClientSet := eventingclientset.NewForConfigOrDie(config)
	//	dynamic := dynamic.NewForConfigOrDie(config)
	//
	//	connector := testkit.NewConnectorClient(
	//		s.testID,
	//		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
	//		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
	//		log.New(),
	//	)
	//	testService := testkit.NewTestService(
	//		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
	//		coreClientset.AppsV1().Deployments(s.testID),
	//		coreClientset.CoreV1().Services(s.testID),
	//		dynamic.Resource(apiRuleRes).Namespace(s.testID),
	//		s.domain,
	//		s.testID,
	//	)
	//
	//	functionEndpoint := helpers.InClusterEndpoint(s.testID, s.testID, helpers.FunctionPort)
	//	state := s.NewState()
	//
	//	return []step.Step{
	//		step.Parallel(
	//			testsuite.NewCreateNamespace(s.testID, coreClientset.CoreV1().Namespaces()),
	//			testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant,
	//				s.applicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
	//				httpSourceClientset.HTTPSources(kymaIntegrationNamespace)),
	//		),
	//		step.Parallel(
	//			testsuite.NewCreateMapping(s.testID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
	//			testsuite.NewDeployFunction(s.testID, helpers.FunctionPayload, helpers.FunctionPort, dynamic.Resource(function).Namespace(s.testID), true),
	//			testsuite.NewStartTestServer(testService),
	//			testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
	//		),
	//		testsuite.NewRegisterTestService(s.testID, testService, state),
	//		testsuite.NewCreateLegacyServiceInstance(s.testID, s.testID, state.GetServiceClassID,
	//			serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
	//			serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
	//		testsuite.NewCreateServiceBinding(s.testID, s.testID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
	//		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
	//			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
	//			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
	//		testsuite.NewCreateKnativeTrigger(s.testID, defaultBrokerName, functionEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
	//		testsuite.NewSleep(s.waitTime),
	//		testsuite.NewSendEventToMesh(s.testID, helpers.FunctionPayload, state),
	//		NewWrappedCounterPod(testService, 1),
	//		testsuite.NewSendEventToCompatibilityLayer(s.testID, helpers.FunctionPayload, state),
	//		NewWrappedCounterPod(testService, 2),
	//	}, nil
}
