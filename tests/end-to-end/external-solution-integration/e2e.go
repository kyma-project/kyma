package main

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var scenarios = map[string]scenario.Scenario{
	"e2e":        &scenario.E2E{},
	"event-only": &scenario.SendEventAndCheckCounter{},
}

var (
	kubeConfig *rest.Config
	runner     *step.Runner
)

func main() {
	scenarioName := os.Args[1]
	os.Args = os.Args[1:]
	scenario, exists := scenarios[scenarioName]
	if !exists {
		log.Errorf("Scenario '%s' does not exist. Use one of the following: ")
		for name := range scenarios {
			log.Infof(" - %s", name)
		}
	}

	runner = step.NewRunner()
	//waitForAPIServer()
	setupLogging()
	setupFlags(scenario)

	steps, err := scenario.Steps(kubeConfig)
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
	time.Sleep(10 * time.Second)
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
