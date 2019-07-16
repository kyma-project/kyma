package scenario

import (
	"fmt"
	"github.com/kyma-project/kyma/common/ingressgateway"
	"github.com/kyma-project/kyma/common/resilient"
	gatewayClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testsuite"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"net/http"
)

type SendEventAndCheckCounter struct {
	domain        string
	runID         string
	testNamespace string
}

type sendEventAndCheckCounterState struct {
	domain string

	registryClient *testkit.RegistryClient
	eventSender    *testkit.EventSender
}

func (s *SendEventAndCheckCounter) AddFlags(set *pflag.FlagSet) {
	pflag.StringVar(&s.testNamespace, "testNamespace", "default", "Namespace where test should create resources")
	pflag.StringVar(&s.domain, "domain", "kyma.local", "domain")
	pflag.StringVar(&s.runID, "runID", "e2e-test", "domain")
}

func (s *SendEventAndCheckCounter) Steps(config *rest.Config) ([]step.Step, error) {
	k8sResourceClient, err := resourceskit.NewK8sResourcesClient(config, s.testNamespace)
	if err != nil {
		return nil, err
	}

	ingressHTTPClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		return nil, err
	}

	gatewayClientset := gatewayClient.NewForConfigOrDie(config)
	connectionTokenHandlerClientset := connectionTokenHandlerClient.NewForConfigOrDie(config)
	tokenRequestClient := resourceskit.NewTokenRequestClient(connectionTokenHandlerClientset.ApplicationconnectorV1alpha1().TokenRequests(s.testNamespace))
	connector := testkit.NewConnectorClient(tokenRequestClient, true, log.New())
	testService := testkit.NewTestService(k8sResourceClient, ingressHTTPClient, gatewayClientset.GatewayV1alpha2(), s.domain)

	state := &sendEventAndCheckCounterState{domain: s.domain}

	return []step.Step{
		testsuite.NewConnectApplication(connector, state),
		testsuite.NewSendEvent(state),
		testsuite.NewCheckCounterPod(testService),
	}, nil
}

func (s *sendEventAndCheckCounterState) SetGatewayHTTPClient(httpClient *http.Client) {
	resilientHttpClient := resilient.WrapHttpClient(httpClient)
	gatewayURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.domain, consts.AppName)
	s.registryClient = testkit.NewRegistryClient(gatewayURL, resilientHttpClient)
	s.eventSender = testkit.NewEventSender(resilientHttpClient, s.domain)
}

func (s *sendEventAndCheckCounterState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}

func (s *sendEventAndCheckCounterState) GetEventSender() *testkit.EventSender {
	return s.eventSender
}
