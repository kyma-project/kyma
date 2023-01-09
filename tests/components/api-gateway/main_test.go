package api_gateway

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestApiGateway(t *testing.T) {
	InitTestSuite()
	SetupCommonResources("api-gateway-tests")
	apiGatewayOpts := goDogOpts
	apiGatewayOpts.Paths = []string{}
	err := filepath.Walk("features/api-gateway", func(path string, info fs.FileInfo, err error) error {
		apiGatewayOpts.Paths = append(apiGatewayOpts.Paths, path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	apiGatewayOpts.Concurrency = conf.TestConcurency
	apigatewaySuite := godog.TestSuite{
		Name:                 "API-Gateway",
		TestSuiteInitializer: InitializeApiGatewayTests,
		Options:              &apiGatewayOpts,
	}

	testExitCode := apigatewaySuite.Run()

	podReport := getPodListReport()
	apiRules := getApiRules()

	//Remove namespace
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err = k8sClient.Resource(res).Delete(context.Background(), namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	if os.Getenv(exportResultVar) == "true" {
		generateReport()
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests, Pod list: %s\n APIRules: %s\n", podReport, apiRules)
	}
}

func InitializeApiGatewayTests(ctx *godog.TestSuiteContext) {
	InitializeScenarioOAuth2Endpoint(ctx.ScenarioContext())
	InitializeScenarioSecuredToUnsecuredEndpoint(ctx.ScenarioContext())
	InitializeScenarioUnsecuredEndpoint(ctx.ScenarioContext())
	InitializeScenarioUnsecuredToSecuredEndpoint(ctx.ScenarioContext())
	InitializeScenarioUnsecuredToSecuredEndpointJWT(ctx.ScenarioContext())
	InitializeScenarioOAuth2JWTOnePath(ctx.ScenarioContext())
	InitializeScenarioOAuth2JWTTwoPaths(ctx.ScenarioContext())
	InitializeScenarioApiruleWithOverrides(ctx.ScenarioContext())
	InitializeScenarioServicePerPath(ctx.ScenarioContext())
}

func InitializeCustomDomainTests(ctx *godog.TestSuiteContext) {
	InitializeScenarioCustomDomain(ctx.ScenarioContext())
}
func TestCustomDomain(t *testing.T) {
	InitTestSuite()
	SetupCommonResources("custom-domain")

	customDomainOpts := goDogOpts
	customDomainOpts.Paths = []string{"features/e2e/custom_domain.feature"}
	customDomainOpts.Concurrency = conf.TestConcurency
	if os.Getenv(exportResultVar) == "true" {
		customDomainOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}
	customDomainSuite := godog.TestSuite{
		Name:                 "custom-domain",
		TestSuiteInitializer: InitializeCustomDomainTests,
		Options:              &customDomainOpts,
	}

	testExitCode := customDomainSuite.Run()

	podReport := getPodListReport()
	apiRules := getApiRules()

	//Remove namespace
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}
	err := k8sClient.Resource(res).Delete(context.Background(), namespace, v1.DeleteOptions{})
	if err != nil {
		log.Print(err.Error())
	}

	if os.Getenv(exportResultVar) == "true" {
		generateReport()
	}

	if testExitCode != 0 {
		t.Fatalf("non-zero status returned, failed to run feature tests, Pod list: %s\n APIRules: %s\n", podReport, apiRules)
	}
}
