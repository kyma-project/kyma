package testsuite

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Config struct {
	Namespace    string        `envconfig:"default=serverless"`
	FunctionName string        `envconfig:"default=test-function"`
	APIRuleName  string        `envconfig:"default=test-apirule"`
	DomainName   string        `envconfig:"default=TODO"`
	DomainHost   string        `envconfig:"default=TODO"`
	DomainPort   uint32        `envconfig:"default=80"`
	WaitTimeout  time.Duration `envconfig:"default=5m"`
}

type TestSuite struct {
	function *function
	apiRule  *apirule.APIRule

	t          *testing.T
	g          *gomega.GomegaWithT
	dynamicCli dynamic.Interface
	cfg        Config
}

func New(restConfig *rest.Config, cfg Config, t *testing.T, g *gomega.GomegaWithT) (*TestSuite, error) {
	_, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	f := newFunction(dynamicCli, cfg.FunctionName, cfg.Namespace, cfg.WaitTimeout, t.Logf)
	ar := apirule.New(dynamicCli, cfg.APIRuleName, cfg.Namespace, cfg.WaitTimeout, t.Logf)

	return &TestSuite{
		function:   f,
		apiRule:    ar,
		t:          t,
		g:          g,
		dynamicCli: dynamicCli,
		cfg:        cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.t.Log("Creating function...")
	functionDetails := t.getFunction()
	resourceVersion, err := t.function.Create(functionDetails, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for APIRule to have ready phase...")
	resourceVersion, err = t.apiRule.Create(t.cfg.DomainName, t.cfg.DomainHost, t.cfg.DomainPort, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Done ;)")

	//TODO: Add ApiRule

	//TODO: Ping function from inside

	//TODO: Ping function from outside
}

func (t *TestSuite) Setup() {
	t.t.Log("Delete old APIRules...")
	err := t.apiRule.Delete(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Delete old functions...")
	err = t.function.Delete(t.t.Log)
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")
	err := t.apiRule.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.function.Delete(t.t.Log)
	failOnError(t.g, err)
}

func (t *TestSuite) getFunction() *functionData {
	return &functionData{
		Body: `module.exports = { main: function(event, context) { return 'Hello World' } }`,
		Deps: `{ "name": "hellowithdeps", "version": "0.0.1", "dependencies": { "end-of-stream": "^1.4.1", "from2": "^2.3.0", "lodash": "^4.17.5" } }`,
	}
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
