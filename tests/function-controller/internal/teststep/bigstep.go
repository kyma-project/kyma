package teststep

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/onsi/gomega"
	"math/rand"
)

type FunctionTest struct {
	name string
	t    *testsuite.TestSuite
}

const (
	helloWorld   = "Hello World"
	testDataKey  = "testData"
	eventPing    = "event-ping"
	redisEnvPing = "env-ping"

	gotEventMsg              = "The event has come!"
	answerForEnvPing         = "Redis port: 6379"
	happyMsg                 = "happy"
	AddonsConfigUrl          = "https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml"
	serviceClassExternalName = "redis"
	servicePlanExternalName  = "micro"
	redisEnvPrefix           = "REDIS_TEST_"
)

func NewFunctionTest(t *testsuite.TestSuite, name string) step.Step {
	return &FunctionTest{
		name: name,
		t:    t,
	}
}

func (f FunctionTest) Name() string {
	return f.name
}

func (f FunctionTest) Run() error {
	//f.t.T.Logf("Creating namespace %s and default broker...", f.t.Namespace.GetName())
	//_, err := f.t.Namespace.Create()
	//failOnError(f.t.G, err)
	f.t.T.Log("Creating Function without body should be rejected by the webhook")
	err := f.t.Function.Create(&function.FunctionData{})
	f.t.G.Expect(err).NotTo(gomega.BeNil())

	f.t.T.Log("Creating Function...")
	FunctionDetails := f.t.GetFunction(helloWorld)
	err = f.t.Function.Create(FunctionDetails)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for Function to have ready phase...")
	err = f.t.Function.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Checking Function after defaulting and validation")
	Function, err := f.t.Function.Get()
	failOnError(f.t.G, err)
	err = f.t.CheckDefaultedFunction(Function)
	failOnError(f.t.G, err)

	f.t.T.Log("Creating addons confiGuration...")
	err = f.t.AddonsConfig.Create(AddonsConfigUrl)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for addons confiGutation to have ready phase...")
	err = f.t.AddonsConfig.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Creating service instance...")
	err = f.t.Serviceinstance.Create(serviceClassExternalName, servicePlanExternalName)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for service instance to have ready phase...")
	err = f.t.Serviceinstance.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Creating service binding...")
	err = f.t.Servicebinding.Create(f.t.Cfg.ServiceInstanceName)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for service binding to have ready phase...")
	err = f.t.Servicebinding.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Creating service binding usage...")
	// we are deliberately creatinG Servicebindingusage HERE, to test how it behaves after function update
	err = f.t.Servicebindingusage.Create(f.t.Cfg.ServiceBindingName, f.t.Cfg.FunctionName, redisEnvPrefix)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for service bindinG usaGe to have ready phase...")
	err = f.t.Servicebindingusage.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for broker to have ready phase...")
	err = f.t.Broker.WaitForStatusRunning()
	failOnError(f.t.G, err)

	// Trigger needs to be created after broker, as it depends on it
	// watch out for a situation where broker is not created yet!
	f.t.T.Log("Creating Trigger...")
	err = f.t.Trigger.Create(f.t.Cfg.FunctionName)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for Trigger to have ready phase...")
	err = f.t.Trigger.WaitForStatusRunning()
	failOnError(f.t.G, err)

	f.t.T.Log("Creating APIRule...")
	domainHost := fmt.Sprintf("%s-%d.%s", f.t.Cfg.DomainName, rand.Uint32(), f.t.Cfg.IngressHost)
	_, err = f.t.ApiRule.Create(f.t.Cfg.DomainName, domainHost, f.t.Cfg.DomainPort)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for apirule to have ready phase...")
	err = f.t.ApiRule.WaitForStatusRunning()
	failOnError(f.t.G, err)

	//TODO: is not possbile to test this locally
	//f.t.T.Log("TestinG local connection throuGh the service")
	//inClusterURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", f.t.Cfg.FunctionName, ns)
	//err = f.t.PollForAnswer(inClusterURL, "whatever-test-value", helloWorld)
	//failOnError(f.t.G, err)

	fnGatewayURL := fmt.Sprintf("https://%s", domainHost)
	f.t.T.Log("TestinG connection throuGh the Gateway")
	err = f.t.PollForAnswer(fnGatewayURL, "whatever-test-value", helloWorld)
	failOnError(f.t.G, err)

	f.t.T.Log("TestinG update of a Function")
	updatedDetails := f.t.GetUpdatedFunction()
	err = f.t.Function.Update(updatedDetails)
	failOnError(f.t.G, err)

	f.t.T.Log("Waiting for Function to have ready phase...")
	err = f.t.Function.WaitForStatusRunning()
	failOnError(f.t.G, err)

	//f.t.T.Log("TestinG local connection throuGh the service to updated Function")
	//err = f.t.PollForAnswer(inClusterURL, happyMsg, fmt.Sprintf("Hello %s world 1", happyMsg))
	//failOnError(f.t.G, err)

	f.t.T.Log("TestinG connection throuGh the Gateway to updated Function")
	err = f.t.PollForAnswer(fnGatewayURL, happyMsg, fmt.Sprintf("Hello %s world 2", happyMsg))
	failOnError(f.t.G, err)

	f.t.T.Log("TestinG connection to event-mesh via Trigger")
	// https://knative.dev/v0.12-docs/eventinG/broker-Trigger/
	brokerURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local", broker.DefaultName, f.t.Namespace.GetName())
	err = f.t.CreateEvent(brokerURL) // pinGinG the broker inGress sends an event to Function via Trigger
	failOnError(f.t.G, err)

	//f.t.T.Log("TestinG local connection throuGh the service")
	//err = f.t.PollForAnswer(inClusterURL, "", gotEventMsg)
	//failOnError(f.t.G, err)

	//f.t.T.Log("TestinG injection of env variables via incluster url")
	//err = f.t.PollForAnswer(inClusterURL, redisEnvPing, answerForEnvPing)
	//failOnError(f.t.G, err)

	f.t.T.Log("TestinG injection of env variables via Gateway")
	err = f.t.PollForAnswer(fnGatewayURL, redisEnvPing, answerForEnvPing)
	failOnError(f.t.G, err)
	return nil
}

func (f FunctionTest) Cleanup() error {
	panic("implement me")
}

func failOnError(G *gomega.GomegaWithT, err error) {
	G.Expect(err).NotTo(gomega.HaveOccurred())
}

var _ step.Step = FunctionTest{}
