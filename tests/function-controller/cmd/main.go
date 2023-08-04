package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite/tests"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
)

func loadRestConfig(context string) (*rest.Config, error) {
	// If the recommended kubeconfig env variable is not specified,
	// try the in-cluster config.
	kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	if len(kubeconfigPath) == 0 {
		if c, err := rest.InClusterConfig(); err == nil {
			return c, nil
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %w", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}

	return loadRestConfigWithContext("", loadingRules, context)
}

func loadRestConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}

type testSuite struct {
	name string
	test test
}

var availableScenarios = map[string][]testSuite{
	"serverless-integration": {
		{name: "simple", test: tests.SimpleFunctionTest},
		{name: "gitops", test: tests.GitopsSteps},
	},
	"git-auth-integration": {{name: "gitauth", test: tests.GitAuthTestSteps}},
	"serverless-contract-tests": {
		{name: "tracing", test: tests.FunctionTracingTest},
		{name: "api-gateway", test: tests.FunctionAPIGatewayTest},
		{name: "cloud-events", test: scenarios.SimpleFunctionCloudEventsTest},
	},
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
		os.Exit(1)
	}

	cfg, err := loadConfig("APP")
	failOnError(err, logf)
	logf.Printf("loaded config")

	scenarioName := os.Args[1]
	logf.Printf("Scenario: %s", scenarioName)
	os.Args = os.Args[1:]
	pickedScenario, exists := availableScenarios[scenarioName]
	if !exists {
		logf.Errorf("Scenario %s not exist", scenarioName)
		os.Exit(2)
	}

	restConfig, err := loadRestConfig("")
	if err != nil {
		logf.Errorf("Unable to get rest config: %s", err.Error())
		os.Exit(3)
	}

	suite := flag.String("test-suite", "", "Choose test-suite to run from scenario")
	flag.Parse()
	if suite != nil && *suite == "" {
		suite = nil
	}
	rand.Seed(time.Now().UnixNano())

	g, _ := errgroup.WithContext(context.Background())
	for _, ts := range pickedScenario {
		if suite != nil && ts.name != *suite {
			logf.Infof("Skip test suite suite: %s", ts.name)
			continue
		}
		// https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		testName := fmt.Sprintf("%s-%s", scenarioName, ts.name)
		func(ts test, name string) {
			g.Go(func() error {
				return runTestSuite(ts, name, logf, cfg, restConfig)
			})
		}(ts.test, testName)
	}
	failOnError(g.Wait(), logf)
}

type test func(*rest.Config, testsuite.Config, *logrus.Entry) (step.Step, error)

func runTestSuite(testToRun test, testSuiteName string, logf *logrus.Logger, cfg config, restConfig *rest.Config) error {
	testSuiteLogger := logf.WithField("test", testSuiteName)
	steps, err := testToRun(restConfig, cfg.Test, testSuiteLogger)
	if err != nil {
		logf.Error(err)
		return err
	}

	runner := step.NewRunner(step.WithCleanupDefault(cfg.Test.Cleanup), step.WithLogger(logf))

	err = runner.Execute(steps)
	if err != nil {
		testSuiteLogger.Error(err)
		return err
	}
	testSuiteLogger.Infof("Test suite succeeded: %s", testSuiteName)
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
