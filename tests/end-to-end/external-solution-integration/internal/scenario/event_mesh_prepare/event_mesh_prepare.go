package event_mesh_prepare

import (
	servicecatalogclient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	eventing "knative.dev/eventing/pkg/client/clientset/versioned"

	appbrokerclient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appoperatorclient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	connectiontokenhandlerclient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	sourcesclientv1alpha1 "github.com/kyma-project/kyma/components/event-sources/client/generated/clientset/internalclientset/typed/sources/v1alpha1"
	servicebindingusageclient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
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
	var (
		apiRuleRes = schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
		function   = schema.GroupVersionResource{Group: "serverless.kyma-project.io", Version: "v1alpha1", Resource: "functions"}
	)

	appOperatorClientset := appoperatorclient.NewForConfigOrDie(config)
	appBrokerClientset := appbrokerclient.NewForConfigOrDie(config)
	coreClientset := coreClient.NewForConfigOrDie(config)
	serviceCatalogClientset := servicecatalogclient.NewForConfigOrDie(config)
	serviceBindingUsageClientset := servicebindingusageclient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectiontokenhandlerclient.NewForConfigOrDie(config)
	httpSourceClientset := sourcesclientv1alpha1.NewForConfigOrDie(config)
	knativeEventingClientSet := eventing.NewForConfigOrDie(config)
	dynamicInterface := dynamic.NewForConfigOrDie(config)

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
		dynamicInterface.Resource(apiRuleRes).Namespace(s.TestID),
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
			testsuite.NewDeployFunction(s.TestID, helpers.FunctionPayload, helpers.FunctionPort, dynamicInterface.Resource(function).Namespace(s.TestID), true),
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
