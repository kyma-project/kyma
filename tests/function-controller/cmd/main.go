package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/scenarios"

	controllerruntime "sigs.k8s.io/controller-runtime"
)

type scenario struct {
	displayName string
	testSuite   testSuite
}

var availableScenarios = map[string][]scenario{
	"serverless-integration": {
		{displayName: "simple", testSuite: scenarios.SimpleFunctionTest},
		{displayName: "gitops", testSuite: scenarios.GitopsSteps}},
	"kyma-integration": {{displayName: "full", testSuite: scenarios.FunctionTestStep}},
}

type config struct {
	Test testsuite.Config
}

func main() {
	logf := logrus.New()
	logf.SetFormatter(&logrus.JSONFormatter{})
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

	g, _ := errgroup.WithContext(context.Background())
	for _, scenario := range pickedScenarios {
		// https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		scenarioDisplayName := fmt.Sprintf("%s-%s", scenarioName, scenario.displayName)
		func(testSuite testSuite, name string) {
			g.Go(func() error {
				return runScenario(testSuite, name, logf, cfg, restConfig)
			})
		}(scenario.testSuite, scenarioDisplayName)
	}
	failOnError(g.Wait(), logf)
}

type testSuite func(*rest.Config, testsuite.Config, *logrus.Entry) (step.Step, error)

func runScenario(testFunc testSuite, name string, logf *logrus.Logger, cfg config, restConfig *rest.Config) error {
	steps, err := testFunc(restConfig, cfg.Test, logf.WithField("suite", name))
	if err != nil {
		logf.Error(err)
		return err
	}

	runner := step.NewRunner(step.WithCleanupDefault(cfg.Test.Cleanup), step.WithLogger(logf))

	err = runner.Execute(steps)
	if err != nil {
		logf.Error(err)
		return err
	}
	return nil
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
