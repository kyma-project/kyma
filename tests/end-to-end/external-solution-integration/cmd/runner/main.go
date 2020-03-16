package main

import (
	"github.com/avast/retry-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/compass_e2e"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/connectivity_adapter_e2e"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/e2e"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_2_phases_prepare"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_2_phases_test"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh_e2e"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	k8s "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/compass"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/connectivity_adapter"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/event_mesh"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario/send_and_check_event"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var scenarios = map[string]scenario.Scenario{
	"event-only":               &send_and_check_event.Scenario{},
	"compass-e2e":              &compass.Scenario{},
	"e2e-event-mesh":           &event_mesh.Scenario{},
	"connectivity-adapter-e2e": &connectivity_adapter.Scenario{},
	"e2e":                      &e2e.E2EScenario{},
	"event-only":               &e2e.SendEventAndCheckCounter{},
	"compass-e2e":              &compass_e2e.CompassE2EScenario{},
	"e2e-event-mesh":           &event_mesh_e2e.E2EEventMeshConfig{},
	"e2e-prepare":              &event_mesh_2_phases_prepare.TwoPhasesEventMeshPrepareConfig{},
	"e2e-test":                 &event_mesh_2_phases_test.TwoPhasesEventMeshTestConfig{},
	"event-mesh":               &event_mesh_e2e.E2EEventMeshConfig{},
	"connectivity-adapter-e2e": &connectivity_adapter_e2e.CompassConnectivityAdapterE2EConfig{},
}

var (
	kubeConfig *rest.Config
	runner     *step.Runner
)

func main() {
	if len(os.Args) < 2 {
		log.Errorf("Scenario not specified. Specify it as the first argument")
		os.Exit(1)
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
