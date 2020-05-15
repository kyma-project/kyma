package event_mesh_prepare

import (
	"time"

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
	CertKey                  = "cert"
	AppNameKey               = "appname"
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
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		log.New(),
	)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.skipSSLVerify)),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		dynamic.Resource(apiRuleRes).Namespace(s.testID),
		s.domain,
		s.testID,
	)

	functionEndpoint := helpers.InClusterEndpoint(s.testID, s.testID, helpers.FunctionPort)
	state := s.NewState()

	dataStore := testkit.NewDataStore(coreClientset, s.testID)

	state.SetDataStore(dataStore)

	return []step.Step{
		step.Parallel(
			testsuite.NewCreateNamespace(s.testID, coreClientset.CoreV1().Namespaces()),
			testsuite.NewCreateApplication(s.testID, s.testID, false, s.applicationTenant,
				s.applicationGroup, appOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
				httpSourceClientset.HTTPSources(kymaIntegrationNamespace)),
		),
		testsuite.NewCreateMapping(s.testID, appBrokerClientset.ApplicationconnectorV1alpha1().ApplicationMappings(s.testID)),
		testsuite.NewDeployFunction(s.testID, helpers.FunctionPayload, helpers.FunctionPort, dynamic.Resource(function).Namespace(s.testID), true),
		testsuite.NewStartTestServer(testService),
		testsuite.NewSleep(5 * time.Second),
		testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
		testsuite.NewRegisterTestService(s.testID, testService, state),
		testsuite.NewCreateLegacyServiceInstance(s.testID, s.testID, state.GetServiceClassID,
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			serviceCatalogClientset.ServicecatalogV1beta1().ServiceClasses(s.testID)),
		testsuite.NewCreateServiceBinding(s.testID, s.testID, serviceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID)),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID,
			serviceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID),
			knativeEventingClientSet.EventingV1alpha1().Brokers(s.testID), knativeEventingClientSet.MessagingV1alpha1().Subscriptions(kymaIntegrationNamespace)),
		testsuite.NewSleep(5 * time.Second),
		testsuite.NewCreateKnativeTrigger(s.testID, defaultBrokerName, functionEndpoint, knativeEventingClientSet.EventingV1alpha1().Triggers(s.testID)),
	}, nil
}
