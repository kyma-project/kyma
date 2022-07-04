package api_gateway

import (
	"context"
	"errors"
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
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/resource"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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
	HydraAddr        string        `envconfig:"TEST_HYDRA_ADDRESS"`
	User             string        `envconfig:"TEST_USER_EMAIL,default=admin@kyma.cx"`
	Pwd              string        `envconfig:"TEST_USER_PASSWORD,default=1234"`
	ReqTimeout       uint          `envconfig:"TEST_REQUEST_TIMEOUT,default=180"`
	ReqDelay         uint          `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain           string        `envconfig:"TEST_DOMAIN"`
	GatewayName      string        `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace string        `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	ClientTimeout    time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` // Don't forget the unit!
	IsMinikubeEnv    bool          `envconfig:"TEST_MINIKUBE_ENV,default=false"`
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

func generateHTMLReport() {
	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: "reports/",
			Title:            "Kyma Istio component tests",
			Metadata: models.Metadata{
				Platform:   runtime.GOOS,
				Parallel:   "Scenarios",
				Executed:   "Remote",
				AppVersion: "main",
				Browser:    "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func initK8sClient() (kubernetes.Interface, dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper) {
	var kubeconfig string
	if kConfig, ok := os.LookupEnv("KUBECONFIG"); !ok {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeconfig = kConfig
	}
	_, err := os.Stat(kubeconfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("kubeconfig %s does not exist", kubeconfig)
		}
		log.Fatalf(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	return k8sClient, dynClient, mapper
}
