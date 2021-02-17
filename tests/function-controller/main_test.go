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
	Test testsuite.Config
}

func TestRuntimes(t *testing.T) {
	runTests(t, scenarios.FunctionTestStep)
}

func TestGitSourcesFunctions(t *testing.T) {
	runTests(t, scenarios.GitopsSteps)
}

type testSuite func(*rest.Config, testsuite.Config, *logrus.Entry) (step.Step, error)

func runTests(t *testing.T, testFunc testSuite) {
	rand.Seed(time.Now().UnixNano())
	g := gomega.NewGomegaWithT(t)

	cfg, err := loadConfig("APP")
	failOnError(g, err)

	restConfig := controllerruntime.GetConfigOrDie()

	logf := logrus.New()
	logf.SetFormatter(&logrus.TextFormatter{})
	logf.SetReportCaller(false)

	steps, err := testFunc(restConfig, cfg.Test, logf.WithField("suite", t.Name()))
	failOnError(g, err)
	runner := step.NewRunner(step.WithCleanupDefault(cfg.Test.Cleanup), step.WithLogger(logf))

	err = runner.Execute(steps)
	failOnError(g, err)
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
