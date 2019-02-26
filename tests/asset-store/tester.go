package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/tests/asset-store/internal/testsuite"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// config contains configuration fields used for upload
type config struct {
	KubeconfigPath string `envconfig:"optional"`
	TestSuite      testsuite.Config
}

// TODO: Transform to Go test
func main() {
	cfg, err := loadConfig("APP")
	parseFlags()
	exitOnError(err, "Error while loading app config")

	restConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while loading K8s REST config")

	testSuite, err := testsuite.New(restConfig, cfg.TestSuite)
	exitOnError(err, "Error while creating test suite")

	err = testSuite.Run()
	exitOnError(err, "Error while running test suite")

	err = testSuite.Cleanup()
	exitOnError(err, "Error while cleaning up after running test suite")
}

func newRestClientConfig(kubeconfigPath string) (*restclient.Config, error) {
	var config *restclient.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = restclient.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}
	return config, nil
}

func parseFlags() {
	err := flag.Set("stderrthreshold", "INFO")
	if err != nil {
		glog.Error(errors.Wrap(err, "while parsing flags"))
	}
	flag.Parse()
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}
