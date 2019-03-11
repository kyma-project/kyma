package main

import (
	"flag"
	"fmt"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/logger"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/signal"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	helloworld "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/hello-world"

	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

// Config holds application configuration
type Config struct {
	Logger              logger.Config
	MaxConcurrencyLevel int `envconfig:"default=1"`
}

const (
	prepareDataActionName  = "prepareData"
	executeTestsActionName = "executeTests"
)

func main() {
	actionUsage := fmt.Sprintf("Define what kind of action runner should execute. Possible values: %s or %s", prepareDataActionName, executeTestsActionName)

	var action string
	flag.StringVar(&action, "action", "", actionUsage)
	flag.Parse()

	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while reading configuration from environment variables")

	log := logrus.New()
	log.SetLevel(logrus.Level(cfg.Logger.Level))

	// Set up signals so we can handle the first shutdown signal gracefully
	stopCh := signal.SetupChannel()

	// Register tests
	tests := []runner.UpgradeTest{
		&helloworld.HelloWorld{},
	}

	// Execute action
	testRunner := runner.NewTestRunner(log, tests, cfg.MaxConcurrencyLevel)
	switch action {
	case prepareDataActionName:
		testRunner.PrepareData(stopCh)
	case executeTestsActionName:
		testRunner.ExecuteTests(stopCh)
	default:
		logrus.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", prepareDataActionName, executeTestsActionName)
	}
}

func fatalOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(fmt.Sprintf("%s: %v", err, context))
	}
}
