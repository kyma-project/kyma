package main

import (
	"math/rand"
	"os"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/scenarios"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/rest"

	"time"

	controllerruntime "sigs.k8s.io/controller-runtime"
)

var availableScenarios = map[string][]testSuite{
	"serverless-integration": {scenarios.SimpleFunctionTest},
	"kyma-integration":       {scenarios.FunctionTestStep, scenarios.GitopsSteps},
}

type config struct {
	Test testsuite.Config
}

func main() {
	logf := logrus.New()
	logf.SetFormatter(&logrus.TextFormatter{})
	logf.SetReportCaller(false)

	if len(os.Args) < 2 {
		logf.Errorf("Scenario not specified. Specify it as the first argument")
		os.Exit(2)
	}

	cfg, err := loadConfig("APP")
	failOnError(err, logf)
	logf.Printf("loaded config")

	restConfig := controllerruntime.GetConfigOrDie()

	scenarioName := os.Args[1]
	logf.Printf("Scenario: %s", scenarioName)
	os.Args = os.Args[1:]
	pickedScenarios, exists := availableScenarios[scenarioName]
	if !exists {
		logf.Errorf("Scenario %s not exist", scenarioName)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	for _, scenario := range pickedScenarios {
		runScenario(scenario, scenarioName, logf, cfg, restConfig)
	}
}

type testSuite func(*rest.Config, testsuite.Config, *logrus.Entry) (step.Step, error)

func runScenario(testFunc testSuite, name string, logf *logrus.Logger, cfg config, restConfig *rest.Config) {
	steps, err := testFunc(restConfig, cfg.Test, logf.WithField("suite", name))
	failOnError(err, logf)
	runner := step.NewRunner(step.WithCleanupDefault(cfg.Test.Cleanup), step.WithLogger(logf))

	err = runner.Execute(steps)
	failOnError(err, logf)
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func failOnError(err error, logf *logrus.Logger) {
	if err != nil {
		logf.Error(err)
		os.Exit(1)
	}
}
