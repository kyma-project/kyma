package istio

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var t *testing.T
var goDogOpts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Format:      "pretty",
	TestingT:    t,
	Concurrency: 1,
}

func InitTest() {
	if os.Getenv(exportResultVar) == "true" {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}
	k8sClient, dynamicClient, mapper = initK8sClient()
}

func TestIstioInstalledEvaluation(t *testing.T) {
	InitTest()

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
	InitTest()

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

func TestIstioReconcilation(t *testing.T) {
	InitTest()

	opts := goDogOpts
	opts.Paths = []string{"features/istio_reconcilation.feature"}

	suite := godog.TestSuite{
		Name:                "reconcilation-tests",
		ScenarioInitializer: InitializeScenarioReconcilation,
		Options:             &opts,
	}

	suiteExitCode := suite.Run()

	if os.Getenv(exportResultVar) == "true" {
		generateHTMLReport()
	}

	if suiteExitCode != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
