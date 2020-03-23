package testsuite

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Config struct {
	Namespace          string        `envconfig:"default=test-function"`
	FunctionName       string        `envconfig:"default=test-function"`
	APIRuleName        string        `envconfig:"default=test-apirule"`
	DomainName         string        `envconfig:"default=test-function"`
	IngressHost        string        `envconfig:"default=kyma.local"`
	DomainPort         uint32        `envconfig:"default=80"`
	InsecureSkipVerify bool          `envconfig:"default=true"`
	WaitTimeout        time.Duration `envconfig:"default=5m"`
}

type TestSuite struct {
	namespace *namespace.Namespace
	function  *function
	apiRule   *apirule.APIRule

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

	ns := namespace.New(coreCli, cfg.Namespace)
	f := newFunction(dynamicCli, cfg.FunctionName, cfg.Namespace, cfg.WaitTimeout, t.Logf)
	ar := apirule.New(dynamicCli, cfg.APIRuleName, cfg.Namespace, cfg.WaitTimeout, t.Logf)

	return &TestSuite{
		namespace:  ns,
		function:   f,
		apiRule:    ar,
		t:          t,
		g:          g,
		dynamicCli: dynamicCli,
		cfg:        cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.t.Log("Creating namespace...")
	err := t.namespace.Create(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Creating function...")
	functionDetails := t.getFunction()
	resourceVersion, err := t.function.Create(functionDetails, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusRunning(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for APIRule to have ready phase...")
	domainHost := fmt.Sprintf("%s.%s", t.cfg.DomainName, t.cfg.IngressHost)
	resourceVersion, err = t.apiRule.Create(t.cfg.DomainName, domainHost, t.cfg.DomainPort, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Testing local connection through the service")
	err = t.checkConnection(fmt.Sprintf("http://%s.%s.svc.cluster.local", t.cfg.FunctionName, t.cfg.Namespace))
	failOnError(t.g, err)

	t.t.Log("Testing connection through the gateway")
	err = t.checkConnection(fmt.Sprintf("https://%s", domainHost))
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")
	err := t.apiRule.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.function.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.namespace.Delete()
	failOnError(t.g, err)
}

func (t *TestSuite) getFunction() *functionData {
	return &functionData{
		Body: `module.exports = { main: function(event, context) { return 'Hello World' } }`,
		Deps: `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "end-of-stream": "^1.4.1", "from2": "^2.3.0", "lodash": "^4.17.5" } }`,
	}
}

func (t *TestSuite) checkConnection(addres string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: t.cfg.InsecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(addres)
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
