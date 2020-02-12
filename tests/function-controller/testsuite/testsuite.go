package testsuite

import (
	"github.com/pkg/errors"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Config struct {
	MockiceName string        `envconfig:"default=rafter-test-svc"`
	WaitTimeout time.Duration `envconfig:"default=3m"`
}

type TestSuite struct {
	testing    *testing.T
	gomega     *gomega.GomegaWithT
	dynamicCli dynamic.Interface
	cfg        Config

	testID string
}

func New(restConfig *rest.Config, cfg Config, t *testing.T, g *gomega.GomegaWithT) (*TestSuite, error) {
	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	return &TestSuite{
		testing:    t,
		gomega:     g,
		dynamicCli: dynamicCli,
		testID:     "singularity",
		cfg:        cfg,
	}, nil
}

func (t *TestSuite) Run() {

}

func (t *TestSuite) Setup() {

}

func (t *TestSuite) Cleanup() {

}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
