package send_and_check_event

import (
	log "github.com/sirupsen/logrus"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	gatewayclientset "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	connectiontokenhandlerclientset "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
)

// Steps return scenario steps
func (s *Scenario) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := k8s.NewForConfigOrDie(config)
	gatewayClientset := gatewayclientset.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectiontokenhandlerclientset.NewForConfigOrDie(config)
	connector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify)),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		gatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.Domain,
		s.testID,
	)
	state := s.NewState()

	return []step.Step{
		testsuite.NewConnectApplication(connector, state, s.ApplicationTenant, s.ApplicationGroup),
		testsuite.NewSendEventToCompatibilityLayer(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 1),
		testsuite.NewSendEventToMesh(s.testID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService, 2),
	}, nil
}
