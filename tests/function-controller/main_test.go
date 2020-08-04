package main

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/internal/scenarios"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

func TestRefactored(t *testing.T) {


	restConfig := controllerruntime.GetConfigOrDie()
	coreCli := corev1.NewForConfigOrDie(restConfig)

	dynamicCli, err := dynamic.NewForConfig(restConfig)

	g := gomega.NewGomegaWithT(t)

	config := &scenarios.Config{
		Log:        t,
		CoreCli:    coreCli,
		DynamicCLI: dynamicCli,
	}

	steps := scenarios.Steps(config)

	runner := step.NewRunner()

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
