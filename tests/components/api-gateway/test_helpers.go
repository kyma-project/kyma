package api_gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gitlab.com/rodrigoodhin/gocure/report/html"

	"github.com/avast/retry-go"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/jwt"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/manifestprocessor"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/resource"
	"github.com/tidwall/pretty"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	testIDLength                   = 8
	OauthClientSecretLength        = 8
	OauthClientIDLength            = 8
	manifestsDirectory             = "manifests/"
	testingAppFile                 = "testing-app.yaml"
	globalCommonResourcesFile      = "global-commons.yaml"
	hydraClientFile                = "hydra-client.yaml"
	noAccessStrategyApiruleFile    = "no_access_strategy.yaml"
	oauthStrategyApiruleFile       = "oauth-strategy.yaml"
	jwtAndOauthStrategyApiruleFile = "jwt-oauth-strategy.yaml"
	jwtAndOauthOnePathApiruleFile  = "jwt-oauth-one-path-strategy.yaml"
	resourceSeparator              = "---"
	defaultHeaderName              = "Authorization"
	exportResultVar                = "EXPORT_RESULT"
	junitFileName                  = "junit-report.xml"
	cucumberFileName               = "cucumber-report.json"
	anyToken                       = "any"
	authorizationHeaderName        = "Authorization"
)

var (
	resourceManager *resource.Manager
	conf            Config
	httpClient      *http.Client
	k8sClient       dynamic.Interface
	helper          *helpers.Helper
	jwtConfig       *jwt.Config
	oauth2Cfg       *clientcredentials.Config
	batch           *resource.Batch
	namespace       string
)

var t *testing.T
var goDogOpts = godog.Options{
	Output:   colors.Colored(os.Stdout),
	Format:   "pretty",
	TestingT: t,
}

type Config struct {
	HydraAddr        string        `envconfig:"TEST_HYDRA_ADDRESS,default=https://oauth2.local.kyma.dev"`
	User             string        `envconfig:"TEST_USER_EMAIL,default=admin@kyma.cx"`
	Pwd              string        `envconfig:"TEST_USER_PASSWORD,default=1234"`
	ReqTimeout       uint          `envconfig:"TEST_REQUEST_TIMEOUT,default=180"`
	ReqDelay         uint          `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain           string        `envconfig:"TEST_DOMAIN,default=local.kyma.dev"`
	GatewayName      string        `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace string        `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	ClientTimeout    time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` // Don't forget the unit!
	IsMinikubeEnv    bool          `envconfig:"TEST_MINIKUBE_ENV,default=false"`
	TestConcurency   int           `envconfig:"TEST_CONCURENCY,default=1"`
}

type Scenario struct {
	namespace   string
	url         string
	apiResource []unstructured.Unstructured
}

type TwoStepScenario struct {
	namespace      string
	url            string
	apiResourceOne []unstructured.Unstructured
	apiResourceTwo []unstructured.Unstructured
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getOAUTHToken(oauth2Cfg clientcredentials.Config) (*oauth2.Token, error) {
	var tokenOAUTH *oauth2.Token
	err := retry.Do(
		func() error {
			token, err := oauth2Cfg.Token(context.Background())
			if err != nil {
				log.Fatalf("Error during Token retrival: %+v", err)
				return err
			}
			tokenOAUTH = token
			return nil
		},
		retry.Delay(500*time.Millisecond), retry.Attempts(3))
	return tokenOAUTH, err
}

func generateReport() {
	htmlOutputDir := "reports/"

	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: htmlOutputDir,
			Title:            "Kyma API-Gateway component tests",
			Metadata: models.Metadata{
				Platform:        runtime.GOOS,
				TestEnvironment: "Gardener GCP",
				Parallel:        "Scenarios",
				Executed:        "Remote",
				AppVersion:      "main",
				Browser:         "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}

	if artifactsDir, ok := os.LookupEnv("ARTIFACTS"); ok {
		filepath.Walk("reports", func(path string, info fs.FileInfo, err error) error {
			copy(path, fmt.Sprintf("%s/report.html", artifactsDir))
			return nil
		})
		copy("./junit-report.xml", fmt.Sprintf("%s/junit-report.xml", artifactsDir))
	}

}

func getApiRules() string {
	res := schema.GroupVersionResource{Group: "gateway.kyma-project.io", Version: "v1alpha1", Resource: "apirules"}
	list, _ := k8sClient.Resource(res).List(context.Background(), v1.ListOptions{})

	toPrint, _ := json.Marshal(list)

	return string(pretty.Pretty(toPrint))
}

func CreateScenario(templateFileName string, namePrefix string) (*Scenario, error) {
	testID := generateRandomString(testIDLength)

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRule, err := manifestprocessor.ParseFromFileWithTemplate(templateFileName, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &Scenario{namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), apiResource: accessRule}, nil
}

func CreateTwoStepScenario(templateFileNameOne string, templateFileNameTwo string, namePrefix string) (*TwoStepScenario, error) {
	testID := generateRandomString(testIDLength)

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRuleOne, err := manifestprocessor.ParseFromFileWithTemplate(templateFileNameOne, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	accessRuleTwo, err := manifestprocessor.ParseFromFileWithTemplate(templateFileNameTwo, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
		GatewayNamespace: conf.GatewayNamespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &TwoStepScenario{namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), apiResourceOne: accessRuleOne, apiResourceTwo: accessRuleTwo}, nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func getPodListReport() string {
	type returnedPodList struct {
		PodList []struct {
			Metadata struct {
				Name              string `json:"name"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	list, _ := k8sClient.Resource(res).Namespace("").List(context.Background(), v1.ListOptions{})

	p := returnedPodList{}
	toMarshal, _ := json.Marshal(list)
	json.Unmarshal(toMarshal, &p)
	toPrint, _ := json.Marshal(p)
	return string(pretty.Pretty(toPrint))
}
