package testsuite

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
)

const (
	helloWorld  = "Hello World"
	testDataKey = "testData"
	eventPing   = "event-ping"
	gotEventMsg = "The event has come!"
	happyMsg    = "happy"
)

type Config struct {
	NamespaceBaseName  string        `envconfig:"default=test-function"`
	FunctionName       string        `envconfig:"default=test-function"`
	APIRuleName        string        `envconfig:"default=test-apirule"`
	DomainName         string        `envconfig:"default=test-function"`
	IngressHost        string        `envconfig:"default=kyma.local"`
	DomainPort         uint32        `envconfig:"default=80"`
	InsecureSkipVerify bool          `envconfig:"default=true"`
	WaitTimeout        time.Duration `envconfig:"default=15m"` // damn istio + knative combo
	Verbose            bool          `envconfig:"default=false"`
	MaxPollingTime     time.Duration `envconfig:"default=5m"`
}

type TestSuite struct {
	namespace  *namespace.Namespace
	function   *function
	apiRule    *apirule.APIRule
	broker     *broker.Broker
	trigger    *trigger.Trigger
	t          *testing.T
	g          *gomega.GomegaWithT
	dynamicCli dynamic.Interface
	cfg        Config
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

	namespaceName := fmt.Sprintf("%s-%d", cfg.NamespaceBaseName, rand.Uint32())

	ns := namespace.New(coreCli, namespaceName, t, cfg.Verbose)
	f := newFunction(dynamicCli, cfg.FunctionName, namespaceName, cfg.WaitTimeout, t, cfg.Verbose)
	ar := apirule.New(dynamicCli, cfg.APIRuleName, namespaceName, cfg.WaitTimeout, t, cfg.Verbose)
	br := broker.New(dynamicCli, namespaceName, cfg.WaitTimeout, t, cfg.Verbose)
	tr := trigger.New(dynamicCli, broker.DefaultName, namespaceName, cfg.WaitTimeout, t, cfg.Verbose)

	return &TestSuite{
		namespace:  ns,
		function:   f,
		apiRule:    ar,
		broker:     br,
		trigger:    tr,
		t:          t,
		g:          g,
		dynamicCli: dynamicCli,
		cfg:        cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.t.Logf("Creating namespace %s and default broker...", t.namespace.GetName())
	ns, err := t.namespace.Create()
	failOnError(t.g, err)

	t.t.Log("Creating function...")
	functionDetails := t.getFunction(helloWorld)
	resourceVersion, err := t.function.Create(functionDetails)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning(resourceVersion)
	failOnError(t.g, err)

	t.t.Log("Waiting for broker to have ready phase...")
	err = t.broker.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Creating trigger...")
	triggerResourceVersion, err := t.trigger.Create(t.cfg.FunctionName)
	failOnError(t.g, err)

	t.t.Log("Creating APIRule...")
	domainHost := fmt.Sprintf("%s-%d.%s", t.cfg.DomainName, rand.Uint32(), t.cfg.IngressHost)
	// var apiruleRsourceVersion string
	_, err = t.apiRule.Create(t.cfg.DomainName, domainHost, t.cfg.DomainPort)
	failOnError(t.g, err)

	t.t.Log("Waiting for trigger to have ready phase...")
	err = t.trigger.WaitForStatusRunning(triggerResourceVersion)
	failOnError(t.g, err)

	t.t.Log("Waiting for apirule to have ready phase...")
	err = t.apiRule.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	inClusterURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", t.cfg.FunctionName, ns)
	t.t.Logf("Address: %s", inClusterURL)
	err = t.pollForAnswer(inClusterURL, helloWorld)
	failOnError(t.g, err)

	fnGatewayURL := fmt.Sprintf("https://%s", domainHost)
	t.t.Log("Testing connection through the gateway")
	t.t.Logf("Address: %s", fnGatewayURL)
	err = t.pollForAnswer(fnGatewayURL, helloWorld)
	failOnError(t.g, err)

	t.t.Log("Testing update of a function")
	updatedDetails := t.getUpdatedFunction(testDataKey, eventPing, gotEventMsg)
	err = t.function.Update(updatedDetails)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning(resourceVersion)
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service to updated function")
	t.t.Logf("Address: %s", inClusterURL)
	err = t.pollForAnswer(inClusterURL, fmt.Sprintf("Hello %s world 1", happyMsg))
	failOnError(t.g, err)

	t.t.Log("Testing connection through the gateway to updated function")
	t.t.Logf("Address: %s", fnGatewayURL)
	err = t.pollForAnswer(fnGatewayURL, fmt.Sprintf("Hello %s world 2", happyMsg))
	failOnError(t.g, err)

	t.t.Log("Testing connection to event-mesh via trigger")
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/
	brokerURL := fmt.Sprintf("http://%s-broker.%s.svc.cluster.local", broker.DefaultName, t.namespace.GetName())
	err = t.createEvent(brokerURL) // pinging the broker ingress sends an event to function via trigger
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	t.t.Logf("Address: %s", inClusterURL)
	err = t.pollForAnswer(inClusterURL, gotEventMsg)
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")
	err := t.apiRule.Delete()
	failOnError(t.g, err)

	err = t.function.Delete()
	failOnError(t.g, err)

	err = t.broker.Delete()
	failOnError(t.g, err)

	err = t.trigger.Delete()
	failOnError(t.g, err)

	err = t.namespace.Delete()
	failOnError(t.g, err)
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func (t *TestSuite) createEvent(url string) error {
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/#manual

	payload := fmt.Sprintf(`{ "testData": "%s" }`, eventPing)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("while creating new request: method %s, url %s, payload %s", http.MethodPost, url, payload)
	}

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

func (t *TestSuite) pollForAnswer(url string, expected string) error {
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
			payload := strings.NewReader(fmt.Sprintf(`{ "%s": "happy" }`, testDataKey))
			req, err := http.NewRequest(http.MethodGet, url, payload)
			if err != nil {
				return true, err
			}

			req.Header.Add("content-type", "application/json")
			res, err := client.Do(req)
			if err != nil {
				return false, err
			}
			defer res.Body.Close()

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
