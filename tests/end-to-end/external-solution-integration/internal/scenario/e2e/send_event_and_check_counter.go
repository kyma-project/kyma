package e2e

import (
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/helpers"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SendEventAndCheckCounter is a shorter version of E2EScenario. It only sends event and checks if counter pod is updated.
type SendEventAndCheckCounter struct {
	E2EScenario
}

// Steps return scenario steps
func (s *SendEventAndCheckCounter) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := coreClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	connector := testkit.NewConnectorClient(
		s.TestID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.TestID),
		internal.NewHTTPClient(s.SkipSSLVerify),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.SkipSSLVerify),
		coreClientset.AppsV1().Deployments(s.TestID),
		coreClientset.CoreV1().Services(s.TestID),
		gatewayClientset.GatewayV1alpha2().Apis(s.TestID),
		s.Domain,
		s.TestID,
	)
	state := s.NewState()

	return []step.Step{
		testsuite.NewConnectApplication(connector, state, s.ApplicationTenant, s.ApplicationGroup),
		testsuite.NewSendEvent(s.TestID, helpers.LambdaPayload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
