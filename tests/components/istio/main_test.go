package istio

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
)

type istioInstalledCase struct {
	pilotPods     *corev1.PodList
	ingressGwPods *corev1.PodList
}

var t *testing.T
var goDogOpts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "pretty",
	TestingT:    t,
	Concurrency: 1,
}

func init() {
	godog.BindCommandLineFlags("godog.", &goDogOpts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	goDogOpts.Paths = pflag.Args()
	k8sClient, dynamicClient, mapper = initK8sClient()
	os.Exit(m.Run())
}

func TestIstioInstalledEvaluation(t *testing.T) {
	evalOpts := goDogOpts
	evalOpts.Paths = []string{"features/istio_evaluation.feature", "features/sidecar_injection.feature"}

	suite := godog.TestSuite{
		Name:                evalProfile,
		ScenarioInitializer: InitializeScenarioIstioInstalled,
		Options:             &evalOpts,
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}
	suiteExitCode := suite.Run()

	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}

	if suiteExitCode != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func TestIstioInstalledProduction(t *testing.T) {
	prodOpts := goDogOpts
	prodOpts.Paths = []string{"features/istio_production.feature", "features/sidecar_injection.feature"}

	suite := godog.TestSuite{
		Name:                prodProfile,
		ScenarioInitializer: InitializeScenarioIstioInstalled,
		Options:             &prodOpts,
	}

	if suite.Name != os.Getenv(deployedKymaProfileVar) {
		t.Skip()
	}

	suiteExitCode := suite.Run()

	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}

	if suiteExitCode != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
