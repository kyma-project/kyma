package main

import (
	"math/rand"
	"testing"
	"time"

	"k8s.io/client-go/rest"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/scenarios"

	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
)

type config struct {
	KubeconfigPath string `envconfig:"optional"`
	Test           testsuite.Config
}

func TestRuntimes(t *testing.T) {
	runTests(t, scenarios.RuntimeSteps, scenarios.RuntimeRun)
}

func TestGitops(t *testing.T) {
	runTests(t, scenarios.GitopsSteps, scenarios.GitopsRun)
}

type testRunner func(*rest.Config, testsuite.Config, *logrus.Logger) ([]step.Step, error)
type testRun func(testsuite.Config) bool

func runTests(t *testing.T, testRunner testRunner, testRun testRun) {
	rand.Seed(time.Now().UnixNano())
	g := gomega.NewGomegaWithT(t)

	cfg, err := loadConfig("APP")
	failOnError(g, err)

	if testRun(cfg.Test) {
		restConfig := controllerruntime.GetConfigOrDie()

		logf := logrus.New()
		logf.SetFormatter(&logrus.TextFormatter{})
		logf.SetReportCaller(false)

		steps, err := testRunner(restConfig, cfg.Test, logf)
		failOnError(g, err)
		runner := step.NewRunner(step.WithCleanupDefault(cfg.Test.Cleanup), step.WithLogger(logf))

		err = runner.Execute(steps)
		failOnError(g, err)
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
