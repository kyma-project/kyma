package main

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/testsuite"
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

type scenario struct {
	displayName string
	scenario    testScenario
}

var availableScenarios = map[string][]scenario{
	"serverless-integration": {
		{displayName: "simple", scenario: testsuite.SimpleFunctionTest},
		{displayName: "gitops", scenario: testsuite.GitopsSteps},
	},
	"git-auth-integration": {{displayName: "gitauth", scenario: testsuite.GitAuthTestSteps}},
	"serverless-contract-tests": {
		{displayName: "tracing", scenario: testsuite.FunctionTracingTest},
		{displayName: "api-gateway", scenario: testsuite.FunctionAPIGatewayTest},
		{displayName: "cloud-events", scenario: testsuite.FunctionCloudEventsTest},
	},
}

type config struct {
	Test internal.Config
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

	scenarioName := os.Args[1]
	logf.Printf("Scenario: %s", scenarioName)
	os.Args = os.Args[1:]
	pickedScenarios, exists := availableScenarios[scenarioName]
	if !exists {
		logf.Errorf("Scenario %s not exist", scenarioName)
		os.Exit(1)
	}

	restConfig, err := loadRestConfig("")
	if err != nil {
		logf.Errorf("Unable to get rest config: %s", err.Error())
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	g, _ := errgroup.WithContext(context.Background())
	for _, scenario := range pickedScenarios {
		// https://eli.thegreenplace.net/2019/go-internals-capturing-loop-variables-in-closures/
		scenarioDisplayName := fmt.Sprintf("%s-%s", scenarioName, scenario.displayName)
		func(scenario testScenario, name string) {
			g.Go(func() error {
				return runScenario(scenario, name, logf, cfg, restConfig)
			})
		}(scenario.scenario, scenarioDisplayName)
	}
	failOnError(g.Wait(), logf)
}

type testScenario func(*rest.Config, internal.Config, *logrus.Entry) (executor.Step, error)

func runScenario(scenario testScenario, scenarioName string, logf *logrus.Logger, cfg config, restConfig *rest.Config) error {
	scenarioLogger := logf.WithField("scenario", scenarioName)
	steps, err := scenario(restConfig, cfg.Test, scenarioLogger)
	if err != nil {
		logf.Error(err)
		return err
	}

	runner := executor.NewRunner(executor.WithCleanupDefault(cfg.Test.Cleanup), executor.WithLogger(logf))

	err = runner.Execute(steps)
	if err != nil {
		scenarioLogger.Error(err)
		return err
	}
	scenarioLogger.Infof("Scenario succeeded: %s", scenarioName)
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
