package main

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"os"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	coreClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var scenarios = map[string]scenario.Scenario{
	"e2e":        &scenario.E2E{},
	"event-only": &scenario.SendEventAndCheckCounter{},
	"compass-e2e": &scenario.CompassE2E{},
}

var (
	kubeConfig *rest.Config
	runner     *step.Runner
)

func main() {
	scenarioName := os.Args[1]
	os.Args = os.Args[1:]
	s, exists := scenarios[scenarioName]
	if !exists {
		log.Errorf("Scenario '%s' does not exist. Use one of the following: ", scenarioName)
		for name := range scenarios {
			log.Infof(" - %s", name)
		}
	}

	runner = step.NewRunner()
	setupLogging()
	setupFlags(s)
	waitForAPIServer()

	steps, err := s.Steps(kubeConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = runner.Execute(steps)

	if err != nil {
		os.Exit(1)
	}

	log.Info("Successfully Finished the e2e test!!")
}

func waitForAPIServer() {
	coreClientset := coreClient.NewForConfigOrDie(kubeConfig)
	err := retry.Do(func() error {
		_, err := coreClientset.CoreV1().Nodes().List(metav1.ListOptions{})
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func setupLogging() {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
}

func setupFlags(s scenario.Scenario) {
	var err error
	kubeconfigFlags := genericclioptions.NewConfigFlags()
	kubeconfigFlags.AddFlags(pflag.CommandLine)
	runner.AddFlags(pflag.CommandLine)
	s.AddFlags(pflag.CommandLine)
	pflag.Parse()

	kubeConfig, err = kubeconfigFlags.ToRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
}
