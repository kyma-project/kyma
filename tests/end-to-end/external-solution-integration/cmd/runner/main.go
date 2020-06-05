package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	k8s "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/compass"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/connectivity_adapter"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_evaluate"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/send_and_check_event"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
)

var scenarios = map[string]scenario.Scenario{
	"event-only":               &send_and_check_event.Scenario{},
	"compass-e2e":              &compass.Scenario{},
	"e2e-event-mesh":           &event_mesh.Scenario{},
	"connectivity-adapter-e2e": &connectivity_adapter.Scenario{},
	"e2e-prepare":              &event_mesh_prepare.Scenario{},
	"e2e-evaluate":             &event_mesh_evaluate.Scenario{},
}

var (
	kubeConfig *rest.Config
	runner     *step.Runner
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Scenario not specified. Specify it as the first argument")
	}

	scenarioName := os.Args[1]
	os.Args = os.Args[1:]
	s, exists := scenarios[scenarioName]
	if !exists {
		log.Errorf("Scenario '%s' does not exist. Use one of the following: ", scenarioName)
		for name := range scenarios {
			log.Infof(" - %s", name)
		}
		os.Exit(1)
	}

	runner = step.NewRunner(s.RunnerOpts()...)
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
	coreClientset := k8s.NewForConfigOrDie(kubeConfig)
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
	kubeconfigFlags := genericclioptions.NewConfigFlags(false)
	kubeconfigFlags.AddFlags(pflag.CommandLine)
	runner.AddFlags(pflag.CommandLine)
	s.AddFlags(pflag.CommandLine)
	pflag.Parse()

	kubeConfig, err = kubeconfigFlags.ToRESTConfig()
	if err != nil {
		log.Fatal(err)
	}
}
