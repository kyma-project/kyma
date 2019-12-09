package main

import (
	"testing"

	"github.com/kyma-project/kyma/tests/asset-store/testsuite"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// config contains configuration fields used for upload
type config struct {
	KubeconfigPath string `envconfig:"optional"`
	Test           testsuite.Config
}

func TestRafter(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	cfg, err := loadConfig("APP")
	failOnError(g, err)

	restConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	failOnError(g, err)

	testSuite, err := testsuite.New(restConfig, cfg.Test, t, g)
	failOnError(g, err)

	testSuite.Run()

	testSuite.Cleanup()
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

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
