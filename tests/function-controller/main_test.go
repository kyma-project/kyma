package main

import (
	"math/rand"
	"testing"
	"time"

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

func TestFunctionController(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	g := gomega.NewGomegaWithT(t)

	cfg, err := loadConfig("APP")
	failOnError(g, err)

	restConfig := controllerruntime.GetConfigOrDie()
	failOnError(g, err)

	testSuite, err := testsuite.New(restConfig, cfg.Test, t, g)
	failOnError(g, err)

	defer testSuite.Cleanup()
	defer testSuite.LogResources()
	testSuite.Run()
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
