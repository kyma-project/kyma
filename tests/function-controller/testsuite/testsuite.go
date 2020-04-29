package testsuite

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
)

type Config struct {
	Namespace          string        `envconfig:"default=test-function"`
	FunctionName       string        `envconfig:"default=test-function"`
	APIRuleName        string        `envconfig:"default=test-apirule"`
	BrokerName         string        `envconfig:"default=default"`
	DomainName         string        `envconfig:"default=test-function"`
	IngressHost        string        `envconfig:"default=kyma.local"`
	DomainPort         uint32        `envconfig:"default=80"`
	InsecureSkipVerify bool          `envconfig:"default=true"`
	WaitTimeout        time.Duration `envconfig:"default=10m"`
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

	namespaceName := fmt.Sprintf("%s-%d", cfg.Namespace, rand.Uint32())

	ns := namespace.New(coreCli, namespaceName, t)
	f := newFunction(dynamicCli, cfg.FunctionName, namespaceName, cfg.WaitTimeout, t)
	ar := apirule.New(dynamicCli, cfg.APIRuleName, namespaceName, cfg.WaitTimeout, t)
	br := broker.New(dynamicCli, namespaceName, cfg.WaitTimeout, t)
	tr := trigger.New(dynamicCli, cfg.BrokerName, namespaceName, cfg.WaitTimeout, t)

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
	t.t.Log("Creating namespace and broker...")
	ns, err := t.namespace.Create()
	failOnError(t.g, err)

	t.t.Log("Creating function...")
	functionDetails := t.getFunction()
	resourceVersion, err := t.function.Create(functionDetails)
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

	t.t.Log("Waiting for broker to have ready phase...")
	err = t.broker.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Waiting for apirule to have ready phase...")
	err = t.apiRule.WaitForStatusRunning()
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning(resourceVersion)
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	err = t.checkConnection(fmt.Sprintf("http://%s.%s.svc.cluster.local", t.cfg.FunctionName, ns))
	failOnError(t.g, err)

	t.t.Log("Testing connection through the gateway")
	err = t.checkConnection(fmt.Sprintf("https://%s", domainHost))
	failOnError(t.g, err)

	t.t.Log("Testing update of a function")
	updatedDetails := t.getUpdatedFunction()
	err = t.function.Update(updatedDetails)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning(resourceVersion)
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	err = t.checkConnection(fmt.Sprintf("http://%s.%s.svc.cluster.local", t.cfg.FunctionName, ns))
	failOnError(t.g, err)

	t.t.Log("Testing connection through the gateway")
	err = t.checkConnection(fmt.Sprintf("https://%s", domainHost))
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

func (t *TestSuite) getFunction() *functionData {
	return &functionData{
		Body: `module.exports = { main: function(event, context) { return 'Hello World' } }`,
		Deps: `{ "name": "hellowithoutdeps", "version": "0.0.1", "dependencies": { } }`,
	}
}

func (t *TestSuite) getUpdatedFunction() *functionData {
	return &functionData{
		Body: `
const _ = require("lodash");

let counter = 0;

module.exports = {
  main: function (event, context) {
    try {
      counter = _.add(counter, 1);
      const eventData = JSON.parse(event.data)["testData"];
      const answer = "Hello " + eventData + " World" + counter;
      console.log(answer);
      return answer;
    } catch (err) {
      console.error(err);
      return "Failed to parse event. Counter value: " + counter;
    }
  }
}
`,
		Deps:        `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "lodash": "^4.17.5" } }`,
		MaxReplicas: 2,
		MinReplicas: 0,
	}
}

func (t *TestSuite) checkConnection(addres string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: t.cfg.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(addres)
	if err == nil {
		defer res.Body.Close()
	}

	if err != nil || res.StatusCode != 200 {
		return errors.Wrapf(err, "while getting response from address %s", addres)
	}

	byteRes, err := ioutil.ReadAll(res.Body)
	if err != nil || string(byteRes) != "Hello World" {
		return errors.Wrap(err, "while reading response")
	}

	return nil
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
