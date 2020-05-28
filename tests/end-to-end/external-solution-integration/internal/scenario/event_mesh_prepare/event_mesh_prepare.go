package event_mesh_prepare

import (
	serviceCatalogClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"

	appBrokerClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appOperatorClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	serviceBindingUsageClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

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

var (
	apiRuleRes = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
	function   = schema.GroupVersionResource{Group: "serverless.kyma-project.io", Version: "v1alpha1", Resource: "functions"}
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	appOperatorClientset := appOperatorClient.NewForConfigOrDie(config)
	appBrokerClientset := appBrokerClient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	serviceCatalogClientset := serviceCatalogClient.NewForConfigOrDie(config)
	serviceBindingUsageClientset := serviceBindingUsageClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	dynamic := dynamic.NewForConfigOrDie(config)

	connector := testkit.NewConnectorClient(
		s.TestID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.TestID),
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		log.New(),
	)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		coreClientset.AppsV1().Deployments(s.TestID),
		coreClientset.CoreV1().Services(s.TestID),
		dynamic.Resource(apiRuleRes).Namespace(s.TestID),
		s.Domain,
		s.TestID,
		s.TestServiceImage,
	)

	functionEndpoint := helpers.InClusterEndpoint(s.TestID, s.TestID, helpers.FunctionPort)
	state := s.NewState()

	dataStore := testkit.NewDataStore(coreClientset, s.TestID)

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.TestID, coreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.TestID, s.TestID, false, s.ApplicationTenant,
				s.ApplicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(kymaIntegrationNamespace)),
		),
		step.Parallel(
			testsuite.NewCreateMapping(s.TestID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.TestID)),
			testsuite.NewDeployFunction(s.TestID, helpers.FunctionPayload, helpers.FunctionPort, dynamic.Resource(function).Namespace(s.TestID), true),
			testsuite.NewStartTestServer(testService),
			testsuite.NewConnectApplication(connector, state, s.ApplicationTenant, s.ApplicationGroup),
		),
		testsuite.NewStoreCertificatesInCluster(dataStore, s.TestID, state.GetCertificates),
		testsuite.NewRegisterTestService(s.TestID, testService, state),
		testsuite.NewCreateLegacyServiceInstance(s.TestID, s.TestID, state.GetServiceClassID,
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.TestID),
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.TestID)),
		testsuite.NewCreateServiceBinding(s.TestID, s.TestID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.TestID)),
		testsuite.NewCreateServiceBindingUsage(s.TestID, s.TestID, s.TestID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.TestID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.TestID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
		testsuite.NewCreateKnativeTrigger(s.TestID, defaultBrokerName, functionEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.TestID)),
	}, nil
}
