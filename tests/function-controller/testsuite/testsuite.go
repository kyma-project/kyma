package testsuite

import (
	"crypto/tls"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/job"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
)

const (
	helloWorld   = "Hello World"
	testDataKey  = "testData"
	eventPing    = "event-ping"
	redisEnvPing = "env-ping"

	gotEventMsg              = "The event has come!"
	answerForEnvPing         = "Redis port: 6379"
	happyMsg                 = "happy"
	addonsConfigUrl          = "https://github.com/kyma-project/addons/releases/download/0.11.0/index-testing.yaml"
	serviceClassExternalName = "redis"
	servicePlanExternalName  = "micro"
	redisEnvPrefix           = "REDIS_TEST_"
)

type Config struct {
	NamespaceBaseName       string        `envconfig:"default=test-function"`
	FunctionName            string        `envconfig:"default=test-function"`
	APIRuleName             string        `envconfig:"default=test-apirule"`
	TriggerName             string        `envconfig:"default=test-trigger"`
	AddonName               string        `envconfig:"default=test-addon"`
	ServiceInstanceName     string        `envconfig:"default=test-service-instance"`
	ServiceBindingName      string        `envconfig:"default=test-service-binding"`
	ServiceBindingUsageName string        `envconfig:"default=test-service-binding-usage"`
	UsageKindName           string        `envconfig:"default=function"`
	DomainName              string        `envconfig:"default=test-function"`
	IngressHost             string        `envconfig:"default=kyma.local"`
	DomainPort              uint32        `envconfig:"default=80"`
	InsecureSkipVerify      bool          `envconfig:"default=true"`
	WaitTimeout             time.Duration `envconfig:"default=15m"` // damn istio + knative combo
	Verbose                 bool          `envconfig:"default=true"`
	MaxPollingTime          time.Duration `envconfig:"default=5m"`
}

type TestSuite struct {
	Namespace           *namespace.Namespace
	Function            *function.Function
	ApiRule             *apirule.APIRule
	Broker              *broker.Broker
	Trigger             *trigger.Trigger
	AddonsConfig        *addons.AddonConfiguration
	Serviceinstance     *serviceinstance.ServiceInstance
	Servicebinding      *servicebinding.ServiceBinding
	Servicebindingusage *servicebindingusage.ServiceBindingUsage
	Jobs                *job.Job
	T                   *testing.T
	G                   *gomega.GomegaWithT
	DynamicCli          dynamic.Interface
	Cfg                 Config
}

func New(restConfig *rest.Config, cfg Config, t *testing.T, g *gomega.GomegaWithT) (*TestSuite, error) {
	coreCli, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s clientset")
	}

	namespaceName := fmt.Sprintf("%s-%d", cfg.NamespaceBaseName, rand.Uint32())

	container := shared.Container{
		DynamicCli:  dynamicCli,
		Namespace:   namespaceName,
		WaitTimeout: cfg.WaitTimeout,
		Verbose:     cfg.Verbose,
		Log:         t,
	}

	ns := namespace.New(namespaceName, coreCli, container)
	f := function.NewFunction(cfg.FunctionName, container)
	ar := apirule.New(cfg.APIRuleName, container)
	br := broker.New(container)
	tr := trigger.New(cfg.TriggerName, container)
	ac := addons.New(cfg.AddonName, container)
	si := serviceinstance.New(cfg.ServiceInstanceName, container)
	sb := servicebinding.New(cfg.ServiceBindingName, container)
	sbu := servicebindingusage.New(cfg.ServiceBindingUsageName, cfg.UsageKindName, container)
	jobList := job.New(cfg.FunctionName, clientset.BatchV1(), container)

	return &TestSuite{
		Namespace:           ns,
		Function:            f,
		ApiRule:             ar,
		Broker:              br,
		Trigger:             tr,
		AddonsConfig:        ac,
		Serviceinstance:     si,
		Servicebinding:      sb,
		Servicebindingusage: sbu,
		Jobs:                jobList,
		T:                   t,
		G:                   g,
		DynamicCli:          dynamicCli,
		Cfg:                 cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.T.Logf("Creating Namespace %s and default Broker...", t.Namespace.GetName())
	ns, err := t.Namespace.Create()
	failOnError(t.G, err)

	t.T.Log("Creating Function without body should be rejected by the webhook")
	err = t.Function.Create(&function.FunctionData{})
	t.G.Expect(err).NotTo(gomega.BeNil())

	//Create function step
	//--------------------------------------------------------------------------------------
	t.T.Log("Creating Function...")
	functionDetails := t.GetFunction(helloWorld)
	err = t.Function.Create(functionDetails)
	failOnError(t.G, err)

	t.T.Log("Waiting for Function to have ready phase...")
	err = t.Function.WaitForStatusRunning()
	failOnError(t.G, err)
	//--------------------------------------------------------------------------------------

	t.T.Log("Checking Function after defaulting and validation")
	function, err := t.Function.Get()
	failOnError(t.G, err)
	err = t.CheckDefaultedFunction(function)
	failOnError(t.G, err)

	t.T.Log("Creating addons configuration...")
	err = t.AddonsConfig.Create(addonsConfigUrl)
	failOnError(t.G, err)

	t.T.Log("Waiting for addons configutation to have ready phase...")
	err = t.AddonsConfig.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Creating service instance...")
	err = t.Serviceinstance.Create(serviceClassExternalName, servicePlanExternalName)
	failOnError(t.G, err)

	t.T.Log("Waiting for service instance to have ready phase...")
	err = t.Serviceinstance.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Creating service binding...")
	err = t.Servicebinding.Create(t.Cfg.ServiceInstanceName)
	failOnError(t.G, err)

	t.T.Log("Waiting for service binding to have ready phase...")
	err = t.Servicebinding.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Creating service binding usage...")
	// we are deliberately creating Servicebindingusage HERE, to test how it behaves after Function update
	err = t.Servicebindingusage.Create(t.Cfg.ServiceBindingName, t.Cfg.FunctionName, redisEnvPrefix)
	failOnError(t.G, err)

	t.T.Log("Waiting for service binding usage to have ready phase...")
	err = t.Servicebindingusage.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Waiting for Broker to have ready phase...")
	err = t.Broker.WaitForStatusRunning()
	failOnError(t.G, err)

	// Trigger needs to be created after Broker, as it depends on it
	// watch out for a situation where Broker is not created yet!
	t.T.Log("Creating Trigger...")
	err = t.Trigger.Create(t.Cfg.FunctionName)
	failOnError(t.G, err)

	t.T.Log("Waiting for Trigger to have ready phase...")
	err = t.Trigger.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Creating APIRule...")
	domainHost := fmt.Sprintf("%s-%d.%s", t.Cfg.DomainName, rand.Uint32(), t.Cfg.IngressHost)
	_, err = t.ApiRule.Create(t.Cfg.DomainName, domainHost, t.Cfg.DomainPort)
	failOnError(t.G, err)

	t.T.Log("Waiting for apirule to have ready phase...")
	err = t.ApiRule.WaitForStatusRunning()
	failOnError(t.G, err)

	t.T.Log("Testing local connection through the service")
	inClusterURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", t.Cfg.FunctionName, ns)
	err = t.PollForAnswer(inClusterURL, "whatever-test-value", helloWorld)
	failOnError(t.G, err)

	fnGatewayURL := fmt.Sprintf("https://%s", domainHost)
	t.T.Log("Testing connection through the gateway")
	err = t.PollForAnswer(fnGatewayURL, "whatever-test-value", helloWorld)
	failOnError(t.G, err)

	t.T.Log("Testing update of a Function")
	updatedDetails := t.GetUpdatedFunction()
	err = t.Function.Update(updatedDetails)
	failOnError(t.G, err)

	t.T.Log("Waiting for Function to have ready phase...")
	err = t.Function.WaitForStatusRunning()
	failOnError(t.G, err)

	//Test Step
	//--------------------------------------------------------------------------------------
	t.T.Log("Testing local connection through the service to updated Function")
	err = t.PollForAnswer(inClusterURL, happyMsg, fmt.Sprintf("Hello %s world 1", happyMsg))
	failOnError(t.G, err)

	t.T.Log("Testing connection through the gateway to updated Function")
	err = t.PollForAnswer(fnGatewayURL, happyMsg, fmt.Sprintf("Hello %s world 2", happyMsg))
	failOnError(t.G, err)

	t.T.Log("Testing connection to event-mesh via Trigger")
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/
	brokerURL := fmt.Sprintf("http://%s-Broker.%s.svc.cluster.local", broker.DefaultName, t.Namespace.GetName())
	err = t.CreateEvent(brokerURL) // pinging the Broker ingress sends an event to Function via Trigger
	failOnError(t.G, err)

	t.T.Log("Testing local connection through the service")
	err = t.PollForAnswer(inClusterURL, "", gotEventMsg)
	failOnError(t.G, err)

	t.T.Log("Testing injection of env variables via incluster url")
	err = t.PollForAnswer(inClusterURL, redisEnvPing, answerForEnvPing)
	failOnError(t.G, err)

	t.T.Log("Testing injection of env variables via gateway")
	err = t.PollForAnswer(fnGatewayURL, redisEnvPing, answerForEnvPing)
	failOnError(t.G, err)
	//--------------------------------------------------------------------------------------
}

func (t *TestSuite) Cleanup() {
	t.T.Log("Cleaning up...")
	err := t.ApiRule.Delete()
	failOnError(t.G, err)

	err = t.Function.Delete()
	failOnError(t.G, err)

	err = t.Trigger.Delete()
	failOnError(t.G, err)

	err = t.Broker.Delete()
	failOnError(t.G, err)

	err = t.Servicebindingusage.Delete()
	failOnError(t.G, err)

	err = t.Servicebinding.Delete()
	failOnError(t.G, err)

	err = t.Serviceinstance.Delete()
	failOnError(t.G, err)

	err = t.AddonsConfig.Delete()
	failOnError(t.G, err)

	err = t.Namespace.Delete()
	failOnError(t.G, err)
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func (t *TestSuite) CreateEvent(url string) error {
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/#manual

	payload := fmt.Sprintf(`{ "%s": "%s" }`, testDataKey, eventPing)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("while creating new request: method %s, url %s, payload %s", http.MethodPost, url, payload)
	}

	// headers taken from example from documentation
	req.Header.Add("x-b3-flags", "1")
	req.Header.Add("ce-specversion", "0.2")
	req.Header.Add("ce-type", "dev.knative.foo.bar")
	req.Header.Add("ce-time", "2018-04-05T03:56:24Z")
	req.Header.Add("ce-id", "45a8b444-3213-4758-be3f-540bf93f85ff")
	req.Header.Add("ce-source", "dev.knative.example")
	req.Header.Add("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while making request to Broker %s", url)
	}

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Invalid response status %s while making a request to %s", resp.Status, url)
	}
	return nil
}

func (t *TestSuite) PollForAnswer(url, payloadStr, expected string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: t.Cfg.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	done := make(chan struct{})

	go func() {
		time.Sleep(t.Cfg.MaxPollingTime)
		close(done)
	}()

	return wait.PollImmediateUntil(10*time.Second,
		func() (done bool, err error) {
			payload := strings.NewReader(fmt.Sprintf(`{ "%s": "%s" }`, testDataKey, payloadStr))
			req, err := http.NewRequest(http.MethodGet, url, payload)
			if err != nil {
				return false, err // TODO erros.wrap
			}

			req.Header.Add("content-type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				return false, err
			}
			defer func() {
				errClose := res.Body.Close()
				if errClose != nil {
					t.T.Logf("Error closing body in request to %s: %v", url, errClose)
				}
			}()

			if res.StatusCode != http.StatusOK {
				t.T.Logf("Expected status %s, got %s, retrying...", http.StatusText(http.StatusOK), res.Status)
				return false, nil
			}

			byteRes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return false, errors.Wrap(err, "while reading response")
			}

			body := string(byteRes)

			if body != expected {
				t.T.Logf("Got: %q, retrying...", body)
				return false, nil
			}

			t.T.Logf("Got: %q, correct...", body)
			return true, nil
		}, done)
}

func (t *TestSuite) LogResources() {
	fn, err := t.Function.Get()
	if err != nil {
		t.T.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	jobs, err := t.Jobs.List()
	if err != nil {
		t.T.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	fnStatusYaml, err := t.prettyYaml(fn.Status)
	if err != nil {
		t.T.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	var jobStatusYaml string
	switch len(jobs.Items) {
	case 0:
		t.T.Log("no job resources matching needed labels")
		jobStatusYaml = "no job resources matching needed labels"
	default:
		jobStatusYaml, err = t.prettyYaml(jobs.Items[0].Status)
		if err != nil {
			t.T.Logf("%v", errors.Wrap(err, "while logging resource"))
		}

	}

	t.T.Logf(`Pretty printed resources:
---
Function's status:
%s
---
Job's status:
%s
---
`, fnStatusYaml, jobStatusYaml)
}

func (t *TestSuite) prettyYaml(resource interface{}) (string, error) {
	out, err := yaml.Marshal(resource)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
