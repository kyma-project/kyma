package scenario

import (
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type SendEventAndCheckCounter struct {
	E2E
}

func (s *SendEventAndCheckCounter) Steps(config *rest.Config) ([]step.Step, error) {
	coreClientset := coreClient.NewForConfigOrDie(config)
	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	tokenRequestClient := resourceskit.NewTokenRequestClient(connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testNamespace))
	connector := testkit.NewConnectorClient(tokenRequestClient, internal.NewHttpClient(s.skipSSLVerify), log.New())
	testService := testkit.NewTestService(
		internal.NewHttpClient(s.skipSSLVerify),
		coreClientset.AppsV1().Deployments(s.testNamespace),
		coreClientset.CoreV1().Services(s.testNamespace),
		gatewayClientset.GatewayV1alpha2().Apis(s.testNamespace),
		s.domain,
		s.testNamespace,
	)
	state := &e2EState{domain: s.domain, skipSSLVerify:s.skipSSLVerify}

	return []step.Step{
		testsuite.NewConnectApplication(connector, state),
		testsuite.NewSendEvent(state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}
