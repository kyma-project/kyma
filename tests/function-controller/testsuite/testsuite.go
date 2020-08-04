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
	namespace           *namespace.Namespace
	function            *function.Function
	apiRule             *apirule.APIRule
	broker              *broker.Broker
	trigger             *trigger.Trigger
	addonsConfig        *addons.AddonConfiguration
	serviceinstance     *serviceinstance.ServiceInstance
	servicebinding      *servicebinding.ServiceBinding
	servicebindingusage *servicebindingusage.ServiceBindingUsage
	jobs                *job.Job
	t                   *testing.T
	g                   *gomega.GomegaWithT
	dynamicCli          dynamic.Interface
	cfg                 Config
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
		namespace:           ns,
		function:            f,
		apiRule:             ar,
		broker:              br,
		trigger:             tr,
		addonsConfig:        ac,
		serviceinstance:     si,
		servicebinding:      sb,
		servicebindingusage: sbu,
		jobs:                jobList,
		t:                   t,
		g:                   g,
		dynamicCli:          dynamicCli,
		cfg:                 cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.t.Logf("Creating namespace %s and default broker...", t.namespace.GetName())
	ns, err := t.namespace.Create()
	failOnError(t.g, err)

	t.t.Log("Creating function without body should be rejected by the webhook")
	err = t.function.Create(&function.FunctionData{})
	t.g.Expect(err).NotTo(gomega.BeNil())

	t.t.Log("Creating function...")
	functionDetails := t.getFunction(helloWorld)
	err = t.function.Create(functionDetails)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Checking function after defaulting and validation")
	function, err := t.function.Get()
	failOnError(t.g, err)
	err = t.checkDefaultedFunction(function)
	failOnError(t.g, err)

	t.t.Log("Creating addons configuration...")
	err = t.addonsConfig.Create(addonsConfigUrl)
	failOnError(t.g, err)

	t.t.Log("Waiting for addons configutation to have ready phase...")
	err = t.addonsConfig.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Creating service instance...")
	err = t.serviceinstance.Create(serviceClassExternalName, servicePlanExternalName)
	failOnError(t.g, err)

	t.t.Log("Waiting for service instance to have ready phase...")
	err = t.serviceinstance.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Creating service binding...")
	err = t.servicebinding.Create(t.cfg.ServiceInstanceName)
	failOnError(t.g, err)

	t.t.Log("Waiting for service binding to have ready phase...")
	err = t.servicebinding.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Creating service binding usage...")
	// we are deliberately creating servicebindingusage HERE, to test how it behaves after function update
	err = t.servicebindingusage.Create(t.cfg.ServiceBindingName, t.cfg.FunctionName, redisEnvPrefix)
	failOnError(t.g, err)

	t.t.Log("Waiting for service binding usage to have ready phase...")
	err = t.servicebindingusage.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Waiting for broker to have ready phase...")
	err = t.broker.WaitForStatusRunning()
	failOnError(t.g, err)

	// trigger needs to be created after broker, as it depends on it
	// watch out for a situation where broker is not created yet!
	t.t.Log("Creating trigger...")
	err = t.trigger.Create(t.cfg.FunctionName)
	failOnError(t.g, err)

	t.t.Log("Waiting for trigger to have ready phase...")
	err = t.trigger.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Creating APIRule...")
	domainHost := fmt.Sprintf("%s-%d.%s", t.cfg.DomainName, rand.Uint32(), t.cfg.IngressHost)
	_, err = t.apiRule.Create(t.cfg.DomainName, domainHost, t.cfg.DomainPort)
	failOnError(t.g, err)

	t.t.Log("Waiting for apirule to have ready phase...")
	err = t.apiRule.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	inClusterURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", t.cfg.FunctionName, ns)
	err = t.pollForAnswer(inClusterURL, "whatever-test-value", helloWorld)
	failOnError(t.g, err)

	fnGatewayURL := fmt.Sprintf("https://%s", domainHost)
	t.t.Log("Testing connection through the gateway")
	err = t.pollForAnswer(fnGatewayURL, "whatever-test-value", helloWorld)
	failOnError(t.g, err)

	t.t.Log("Testing update of a function")
	updatedDetails := t.getUpdatedFunction()
	err = t.function.Update(updatedDetails)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service to updated function")
	err = t.pollForAnswer(inClusterURL, happyMsg, fmt.Sprintf("Hello %s world 1", happyMsg))
	failOnError(t.g, err)

	t.t.Log("Testing connection through the gateway to updated function")
	err = t.pollForAnswer(fnGatewayURL, happyMsg, fmt.Sprintf("Hello %s world 2", happyMsg))
	failOnError(t.g, err)

	t.t.Log("Testing connection to event-mesh via trigger")
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/
	brokerURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local", broker.DefaultName, t.namespace.GetName())
	err = t.createEvent(brokerURL) // pinging the broker ingress sends an event to function via trigger
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	err = t.pollForAnswer(inClusterURL, "", gotEventMsg)
	failOnError(t.g, err)

	t.t.Log("Testing injection of env variables via incluster url")
	err = t.pollForAnswer(inClusterURL, redisEnvPing, answerForEnvPing)
	failOnError(t.g, err)

	t.t.Log("Testing injection of env variables via gateway")
	err = t.pollForAnswer(fnGatewayURL, redisEnvPing, answerForEnvPing)
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")
	err := t.apiRule.Delete()
	failOnError(t.g, err)

	err = t.function.Delete()
	failOnError(t.g, err)

	err = t.trigger.Delete()
	failOnError(t.g, err)

	err = t.broker.Delete()
	failOnError(t.g, err)

	err = t.servicebindingusage.Delete()
	failOnError(t.g, err)

	err = t.servicebinding.Delete()
	failOnError(t.g, err)

	err = t.serviceinstance.Delete()
	failOnError(t.g, err)

	err = t.addonsConfig.Delete()
	failOnError(t.g, err)

	err = t.namespace.Delete()
	failOnError(t.g, err)
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func (t *TestSuite) createEvent(url string) error {
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
		return errors.Wrapf(err, "while making request to broker %s", url)
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

func (t *TestSuite) pollForAnswer(url, payloadStr, expected string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: t.cfg.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}

	done := make(chan struct{})

	go func() {
		time.Sleep(t.cfg.MaxPollingTime)
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
					t.t.Logf("Error closing body in request to %s: %v", url, errClose)
				}
			}()

			if res.StatusCode != http.StatusOK {
				t.t.Logf("Expected status %s, got %s, retrying...", http.StatusText(http.StatusOK), res.Status)
				return false, nil
			}

			byteRes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return false, errors.Wrap(err, "while reading response")
			}

			body := string(byteRes)

			if body != expected {
				t.t.Logf("Got: %q, retrying...", body)
				return false, nil
			}

			t.t.Logf("Got: %q, correct...", body)
			return true, nil
		}, done)
}

func (t *TestSuite) LogResources() {
	fn, err := t.function.Get()
	if err != nil {
		t.t.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	jobs, err := t.jobs.List()
	if err != nil {
		t.t.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	fnStatusYaml, err := t.prettyYaml(fn.Status)
	if err != nil {
		t.t.Logf("%v", errors.Wrap(err, "while logging resource"))
	}

	var jobStatusYaml string
	switch len(jobs.Items) {
	case 0:
		t.t.Log("no job resources matching needed labels")
		jobStatusYaml = "no job resources matching needed labels"
	default:
		jobStatusYaml, err = t.prettyYaml(jobs.Items[0].Status)
		if err != nil {
			t.t.Logf("%v", errors.Wrap(err, "while logging resource"))
		}

	}

	t.t.Logf(`Pretty printed resources:
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
