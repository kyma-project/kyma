package testsuite

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/namespace"

	"github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Config struct {
	Namespace    string        `envconfig:"default=serverless"`
	FunctionName string        `envconfig:"default=test-function"`
	WaitTimeout  time.Duration `envconfig:"default=3m"`
}

type TestSuite struct {
	function  *function
	namespace *namespace.Namespace

	t          *testing.T
	g          *gomega.GomegaWithT
	dynamicCli dynamic.Interface
	cfg        Config

	testID string
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

	return &TestSuite{
		function:   f,
		namespace:  ns,
		t:          t,
		g:          g,
		dynamicCli: dynamicCli,
		testID:     "singularity",
		cfg:        cfg,
	}, nil
}

func (t *TestSuite) Run() {
	t.t.Log("Creating namespace...")
	err := t.namespace.Create(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Creating function...")
	resourceVersion, err := t.function.Create("","","",t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for function to have ready phase...")
	err = t.function.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Done ;)")
	//TODO: Ping function from inside

	//TODO: Ping function from outside
}

func (t *TestSuite) Setup() {
	t.t.Log("Delete old functions...")
	err := t.function.Delete(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Delete old namespace...")
	err = t.namespace.Delete(t.t.Log)
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")
	err := t.function.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.namespace.Delete(t.t.Log)
	failOnError(t.g, err)
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
