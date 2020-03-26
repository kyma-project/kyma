package event_mesh

import (
	kubelessclientset "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	servicecatalogclientset "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	log "github.com/sirupsen/logrus"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	eventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"

	gatewayclientset "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	appbrokerclientset "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appoperatorclientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectiontokenhandlerclientset "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	sbuclientset "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

const (
	kymaIntegrationNamespace = "kyma-integration"
	defaultBrokerName        = "default"
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	appOperatorClientset := appoperatorclientset.NewForConfigOrDie(config)
	appBrokerClientset := appbrokerclientset.NewForConfigOrDie(config)
	kubelessClientset := kubelessclientset.NewForConfigOrDie(config)
	coreClientset := k8s.NewForConfigOrDie(config)
	pods := coreClientset.CoreV1().Pods(s.testID)
	serviceCatalogClientset := servicecatalogclientset.NewForConfigOrDie(config)
	serviceBindingUsageClientset := sbuclientset.NewForConfigOrDie(config)
	gatewayClientset := gatewayclientset.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectiontokenhandlerclientset.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	knativeEventingClientSet := eventingclientset.NewForConfigOrDie(config)

	connector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		gatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, helpers.LambdaPort)
	state := s.NewState()

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, coreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant,
				s.applicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(kymaIntegrationNamespace)),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.testID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
			testsuite.NewDeployLambda(s.testID, helpers.LambdaPayload, helpers.LambdaPort, kubelessClientset.KubelessV1beta1().Functions(s.testID), pods, true),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
		),
		testsuite.NewRegisterTestService(s.testID, testService, state),
		testsuite.NewCreateLegacyServiceInstance(s.testID, s.testID, state.GetServiceClassID,
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
		testsuite.NewCreateServiceBinding(s.testID, s.testID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.testID, defaultBrokerName, lambdaEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
		testsuite.NewSendEventToMesh(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 1),
		testsuite.NewSendEventToCompatibilityLayer(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 2),
	}, nil
}
