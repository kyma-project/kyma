package api_gateway

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/client"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/resource"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/cucumber/godog"

	"github.com/kyma-project/kyma/common/ingressgateway"

	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/manifestprocessor"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func InitApiGatewayTest() {
	pflag.Parse()
	goDogOpts.Paths = pflag.Args()

	if os.Getenv(exportResultVar) == "true" {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	if err := envconfig.Init(&conf); err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}

	if conf.IsMinikubeEnv {
		var err error
		log.Printf("Using dedicated ingress client")
		httpClient, err = ingressgateway.FromEnv().Client()
		if err != nil {
			log.Fatalf("Unable to initialize ingress gateway client: %v", err)
		}
	} else {
		log.Printf("Fallback to default http client")
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: conf.ClientTimeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	oauthClientID := generateRandomString(OauthClientIDLength)
	oauthClientSecret := generateRandomString(OauthClientSecretLength)
	namespace = fmt.Sprintf("api-gateway-test-%s", generateRandomString(6))
	randomSuffix6 := generateRandomString(6)
	oauthSecretName := fmt.Sprintf("api-gateway-tests-secret-%s", randomSuffix6)
	oauthClientName := fmt.Sprintf("api-gateway-tests-client-%s", randomSuffix6)
	log.Printf("Using namespace: %s\n", namespace)
	log.Printf("Using OAuth2Client with name: %s, secretName: %s\n", oauthClientName, oauthSecretName)

	oauth2Cfg = &clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", conf.HydraAddr),
		Scopes:       []string{"read"},
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	jwtConf, err := jwt.LoadConfig(oauthClientID)
	if err != nil {
		log.Fatal(err)
	}

	jwtConfig = &jwtConf

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(conf.ReqDelay) * time.Second),
		retry.Attempts(conf.ReqTimeout / conf.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	}

	helper = helpers.NewHelper(httpClient, commonRetryOpts)

	mapper, err := client.GetDiscoveryMapper()
	if err != nil {
		t.Fatal(err)
	}

	client, err := client.GetDynamicClient()
	if err != nil {
		t.Fatal(err)
	}
	k8sClient = client
	resourceManager = &resource.Manager{RetryOptions: commonRetryOpts}

	batch = &resource.Batch{
		ResourceManager: resourceManager,
		Mapper:          mapper,
	}

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate(globalCommonResourcesFile, manifestsDirectory, resourceSeparator, struct {
		Namespace         string
		OauthClientSecret string
		OauthClientID     string
		OauthSecretName   string
	}{
		Namespace:         namespace,
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(oauthClientSecret)),
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(oauthClientID)),
		OauthSecretName:   oauthSecretName,
	})
	if err != nil {
		log.Fatal(err)
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := batch.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	err = resourceManager.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	_, err = batch.CreateResources(k8sClient, globalCommonResources...)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate(hydraClientFile, manifestsDirectory, resourceSeparator, struct {
		Namespace       string
		OauthClientName string
		OauthSecretName string
	}{
		Namespace:       namespace,
		OauthClientName: oauthClientName,
		OauthSecretName: oauthSecretName,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating hydra client resources")

	_, err = batch.CreateResources(k8sClient, hydraClientResource...)
	if err != nil {
		log.Fatal(err)
	}

	// Let's wait a bit to register client in hydra
	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	// Get HydraClient Status
	hydraClientResourceSchema, ns, name := batch.GetResourceSchemaAndNamespace(hydraClientResource[0])
	clientStatus, err := resourceManager.GetStatus(k8sClient, hydraClientResourceSchema, ns, name)
	errorStatus, ok := clientStatus["reconciliationError"].(map[string]interface{})
	if err != nil || !ok {
		t.Fatalf("Error retrieving Oauth2Client status: %+v | %+v", err, ok)
	}
	if len(errorStatus) != 0 {
		t.Fatalf("Invalid status in Oauth2Client resource: %+v", errorStatus)
	}
}

func InitCustomDomainTest() {
	pflag.Parse()
	goDogOpts.Paths = pflag.Args()

	if os.Getenv(exportResultVar) == "true" {
		goDogOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	if err := envconfig.Init(&conf); err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}

	if conf.IsMinikubeEnv {
		var err error
		log.Printf("Using dedicated ingress client")
		httpClient, err = ingressgateway.FromEnv().Client()
		if err != nil {
			log.Fatalf("Unable to initialize ingress gateway client: %v", err)
		}
	} else {
		log.Printf("Fallback to default http client")
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: conf.ClientTimeout,
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	namespace = fmt.Sprintf("custom-domain-test-%s", generateRandomString(6))
	log.Printf("Using namespace: %s\n", namespace)

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(conf.ReqDelay) * time.Second),
		retry.Attempts(conf.ReqTimeout / conf.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	}

	helper = helpers.NewHelper(httpClient, commonRetryOpts)

	client, err := client.GetDynamicClient()
	if err != nil {
		t.Fatal(err)
	}

	k8sClient = client
	resourceManager = &resource.Manager{RetryOptions: commonRetryOpts}

	batch = &resource.Batch{
		ResourceManager: resourceManager,
	}

}

func TestApiGateway(t *testing.T) {
	InitApiGatewayTest()
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

func TestCustomDomain(t *testing.T) {
	InitApiGatewayTest()
	loadBalancerIP, _ := getLoadBalancerIP()
	fmt.Println(loadBalancerIP)
	customDomainOpts := goDogOpts
	customDomainOpts.Paths = []string{"features/gardener/custom_domain.feature"}
	customDomainOpts.Concurrency = conf.TestConcurency

	if os.Getenv(exportResultVar) == "true" {
		customDomainOpts.Format = "pretty,junit:junit-report.xml,cucumber:cucumber-report.json"
	}

	customDomainSuite := godog.TestSuite{
		Name:                "custom-domain",
		ScenarioInitializer: InitializeScenarioCustomDomain,
		Options:             &customDomainOpts,
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
