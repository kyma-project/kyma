package scenario

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
)

// CompassE2E executes complete external solution integration test scenario
// using Compass for Application registration and connectivity
type CompassE2E struct {
	domain            string
	testID            string
	skipSSLVerify     bool
	applicationTenant string
	applicationGroup  string
	lambdaPort        int
}

// AddFlags adds CLI flags to given FlagSet
func (s *CompassE2E) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.testID, "testID", "compass-e2e-test", "domain")
	pflag.BoolVar(&s.skipSSLVerify, "skipSSLVerify", false, "Skip verification of service SSL certificates")
	pflag.StringVar(&s.applicationTenant, "applicationTenant", "", "Application CR Tenant")
	pflag.StringVar(&s.applicationGroup, "applicationGroup", "", "Application CR Group")
	pflag.IntVar(&s.lambdaPort, "lambdaPort", 8080, "Lambda port")
}

// Steps return scenario steps
func (s *CompassE2E) Steps(config *rest.Config) ([]step.Step, error) {
	state, err := s.NewState()
	if err != nil {
		return nil, err
	}

	kymaClients := testkit.InitKymaClients(config, s.testID)
	compassClients := testkit.InitCompassClients(kymaClients, state, s.domain, s.skipSSLVerify)

	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		kymaClients.CoreClientset.AppsV1().Deployments(s.testID),
		kymaClients.CoreClientset.CoreV1().Services(s.testID),
		kymaClients.GatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)

	lambdaEndpoint := helpers.LambdaInClusterEndpoint(s.testID, s.testID, s.lambdaPort)

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
			testsuite.NewDeployLambda(s.testID, payload, s.lambdaPort, kymaClients.KubelessClientset.KubelessV1beta1().Functions(s.testID), kymaClients.Pods),
			testsuite.NewConnectApplicationUsingCompass(compassClients.ConnectorClient, compassClients.DirectorClient, state),
		),
		testsuite.NewCreateSeparateServiceInstance(s.testID,
			kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceInstances(s.testID),
			kymaClients.AppOperatorClientset.ApplicationconnectorV1alpha1().Applications(),
			state,
		),
		testsuite.NewCreateServiceBinding(s.testID, kymaClients.ServiceCatalogClientset.ServicecatalogV1beta1().ServiceBindings(s.testID), state),
		testsuite.NewCreateServiceBindingUsage(s.testID, s.testID, s.testID, kymaClients.ServiceBindingUsageClientset.ServicecatalogV1alpha1().ServiceBindingUsages(s.testID), nil, nil),
		testsuite.NewCreateSubscription(s.testID, s.testID, lambdaEndpoint, kymaClients.EventingClientset.EventingV1alpha1().Subscriptions(s.testID)),
		testsuite.NewSendEvent(s.testID, payload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
