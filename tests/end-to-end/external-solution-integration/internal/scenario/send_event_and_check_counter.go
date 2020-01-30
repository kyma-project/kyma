package scenario

import (
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// SendEventAndCheckCounter is a shorter version of E2E. It only sends event and checks if counter pod is updated.
type SendEventAndCheckCounter struct {
	E2E
}

// Steps return scenario steps
func (s *SendEventAndCheckCounter) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := coreClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	connector := testkit.NewConnectorClient(
		s.testID,
		connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testID),
		internal.NewHTTPClient(s.skipSSLVerify),
		log.New(),
	)
	testService := testkit.NewTestService(
		internal.NewHTTPClient(s.skipSSLVerify),
		coreClientset.AppsV1().Deployments(s.testID),
		coreClientset.CoreV1().Services(s.testID),
		gatewayClientset.GatewayV1alpha2().Apis(s.testID),
		s.domain,
		s.testID,
	)
	state := s.NewState()

	return []step.Step{
		testsuite.NewConnectApplication(connector, state, s.applicationTenant, s.applicationGroup),
		testsuite.NewSendEvent(s.testID, payload, state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
