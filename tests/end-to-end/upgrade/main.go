package main

import (
	"flag"
	"fmt"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"

	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	k8sclientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	ab "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	ao "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/logger"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/platform/signal"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	monitoring "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/monitoring"
	servicecatalog "github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/tests/service-catalog"
)

// Config holds application configuration
type Config struct {
	Logger              logger.Config
	MaxConcurrencyLevel int    `envconfig:"default=1"`
	KubeconfigPath      string `envconfig:"optional"`
}

const (
	prepareDataActionName  = "prepareData"
	executeTestsActionName = "executeTests"
)

func main() {
	actionUsage := fmt.Sprintf("Define what kind of action runner should execute. Possible values: %s or %s", prepareDataActionName, executeTestsActionName)

	var action string
	var verbose bool
	flag.StringVar(&action, "action", "", actionUsage)
	flag.BoolVar(&verbose, "verbose", false, "Print all test logs")
	flag.Parse()

	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while reading configuration from environment variables")

	log := logger.New(&cfg.Logger)

	// Set up signals so we can handle the first shutdown signal gracefully
	stopCh := signal.SetupChannel()

	// K8s client
	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	fatalOnError(err, "while creating k8s client cfg")

	k8sCli, err := k8sclientset.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating k8s clientset")

	scCli, err := sc.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating Service CAtalog clientset")

	buCli, err := bu.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating Binding Usage clientset")

	appConnectorCli, err := ao.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating Application Connector clientset")

	appBrokerCli, err := ab.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating Application Broker clientset")

	// Register tests. Convention:
	// <test-name> : <test-instance>

	// Using map is on purpose - we ensure that test name will not be duplicated.
	// Test name is sanitized and used for creating dedicated namespace for given test,
	// so it cannot overlap with others.

	grafanaUpgradeTest := monitoring.NewGrafanaUpgradeTest(k8sCli)

	metricUpgradeTest, err := monitoring.NewMetricsUpgradeTest(k8sCli)
	fatalOnError(err, "while creating Metrics Upgrade Test")

	tests := map[string]runner.UpgradeTest{
		"HelmBrokerUpgradeTest":        servicecatalog.NewHelmBrokerTest(k8sCli, scCli, buCli),
		"ApplicationBrokerUpgradeTest": servicecatalog.NewAppBrokerUpgradeTest(scCli, k8sCli, buCli, appBrokerCli, appConnectorCli),
		"GrafanaUpgradeTest":           grafanaUpgradeTest,
		"MetricsUpgradeTest":           metricUpgradeTest,
	}

	// Execute requested action
	testRunner, err := runner.NewTestRunner(log, k8sCli.CoreV1().Namespaces(), tests, cfg.MaxConcurrencyLevel, verbose)
	fatalOnError(err, "while creating test runner")

	switch action {
	case prepareDataActionName:
		err := testRunner.PrepareData(stopCh)
		fatalOnError(err, "while executing prepare data for all registered tests")
	case executeTestsActionName:
		err := testRunner.ExecuteTests(stopCh)
		fatalOnError(err, "while executing tests for all registered tests")
	default:
		logrus.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", prepareDataActionName, executeTestsActionName)
	}
}

func fatalOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(fmt.Sprintf("%s: %v", context, err))
	}
}

func newRestClientConfig(kubeConfigPath string) (*restclient.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return restclient.InClusterConfig()
}
