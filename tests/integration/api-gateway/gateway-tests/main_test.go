package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/client"
	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/resource"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/common/ingressgateway"

	"github.com/avast/retry-go"

	"github.com/stretchr/testify/assert"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/api"
	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/jwt"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/manifestprocessor"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const testIDLength = 8
const OauthClientSecretLength = 8
const OauthClientIDLength = 8
const manifestsDirectory = "manifests/"
const testingAppFile = "testing-app.yaml"
const globalCommonResourcesFile = "global-commons.yaml"
const hydraClientFile = "hydra-client.yaml"
const noAccessStrategyApiruleFile = "no_access_strategy.yaml"
const oauthStrategyApiruleFile = "oauth-strategy.yaml"
const jwtAndOauthStrategyApiruleFile = "jwt-oauth-strategy.yaml"
const jwtAndOauthOnePathApiruleFile = "jwt-oauth-one-path-strategy.yaml"
const resourceSeparator = "---"
const defaultHeaderName = "Authorization"

var (
	resourceManager *resource.Manager
	conf            Config
	httpClient      *http.Client
)

type Config struct {
	HydraAddr        string        `envconfig:"TEST_HYDRA_ADDRESS"`
	User             string        `envconfig:"TEST_USER_EMAIL"`
	Pwd              string        `envconfig:"TEST_USER_PASSWORD"`
	ReqTimeout       uint          `envconfig:"TEST_REQUEST_TIMEOUT,default=180"`
	ReqDelay         uint          `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain           string        `envconfig:"TEST_DOMAIN"`
	GatewayName      string        `envconfig:"TEST_GATEWAY_NAME,default=kyma-gateway"`
	GatewayNamespace string        `envconfig:"TEST_GATEWAY_NAMESPACE,default=kyma-system"`
	ClientTimeout    time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` //Don't forget the unit!
	IsMinikubeEnv    bool          `envconfig:"TEST_MINIKUBE_ENV,default=false"`
}

func TestApiGatewayIntegration(t *testing.T) {

	assert, require := assert.New(t), require.New(t)

	if err := envconfig.Init(&conf); err != nil {
		t.Fatalf("Unable to setup config: %v", err)
	}

	if conf.IsMinikubeEnv {
		var err error
		log.Printf("Using dedicated ingress client")
		httpClient, err = ingressgateway.FromEnv().Client()
		if err != nil {
			t.Fatalf("Unable to initialize ingress gateway client: %v", err)
		}
	} else {
		log.Printf("Fallback to default http client")
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: conf.ClientTimeout,
		}
	}

	oauthClientID := generateRandomString(OauthClientIDLength)
	oauthClientSecret := generateRandomString(OauthClientSecretLength)
	namespace := fmt.Sprintf("api-gateway-test-%s", generateRandomString(6))
	randomSuffix6 := generateRandomString(6)
	oauthSecretName := fmt.Sprintf("api-gateway-tests-secret-%s", randomSuffix6)
	oauthClientName := fmt.Sprintf("api-gateway-tests-client-%s", randomSuffix6)
	log.Printf("Using namespace: %s\n", namespace)
	log.Printf("Using OAuth2Client with name: %s, secretName: %s\n", oauthClientName, oauthSecretName)

	oauth2Cfg := clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", conf.HydraAddr),
		Scopes:       []string{"read"},
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	jwtConfig, err := jwt.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	commonRetryOpts := []retry.Option{
		retry.Delay(time.Duration(conf.ReqDelay) * time.Second),
		retry.Attempts(conf.ReqTimeout / conf.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	}

	tester := api.NewTester(httpClient, commonRetryOpts)

	k8sClient := client.GetDynamicClient()
	resourceManager = &resource.Manager{RetryOptions: commonRetryOpts}

	batch := &resource.Batch{
		resourceManager,
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
		panic(err)
	}

	// delete test namespace if the previous test namespace persists
	nsResourceSchema, ns, name := resource.GetResourceSchemaAndNamespace(globalCommonResources[0])
	log.Printf("Delete test namespace, if exists: %s\n", name)
	resourceManager.DeleteResource(k8sClient, nsResourceSchema, ns, name)

	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	log.Printf("Creating common tests resources")
	batch.CreateResources(k8sClient, globalCommonResources...)
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
		panic(err)
	}
	log.Printf("Creating hydra client resources")
	batch.CreateResources(k8sClient, hydraClientResource...)
	// Let's wait a bit to register client in hydra
	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)
	// Get HydraClient Status
	hydraClientResourceSchema, ns, name := resource.GetResourceSchemaAndNamespace(hydraClientResource[0])
	clientStatus, err := resourceManager.GetStatus(k8sClient, hydraClientResourceSchema, ns, name)
	errorStatus, ok := clientStatus["reconciliationError"].(map[string]interface{})
	if err != nil || !ok {
		t.Fatalf("Error retrieving Oauth2Client status: %+v | %+v", err, ok)
	}
	if len(errorStatus) != 0 {
		t.Fatalf("Invalid status in Oauth2Client resource: %+v", errorStatus)
	}
	// defer deleting namespace (it will also delete all remaining resources in that namespace)
	defer func() {
		time.Sleep(time.Second * 3)
		resourceManager.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	}()

	t.Run("API Gateway should", func(t *testing.T) {
		t.Run("Expose service without access strategy (plain access)", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Forward client IP", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			var clientIP string
			clientIP, err = lookupIPOfTestPod()
			if err != nil {
				t.Fatalf("could not determine ip of running pod for test %s, details %s", t.Name(), err.Error())
			}
			verifyIP := func(receivedContent string) bool {
				ipFound := strings.Contains(receivedContent, clientIP)
				if !ipFound {
					log.Printf("client ip not found in response (%s, %s)", clientIP, receivedContent)
				}
				return ipFound
			}

			assert.NoError(tester.TestUnsecuredEndpointContent(fmt.Sprintf("https://httpbin-%s.%s/ip", testID, conf.Domain), verifyIP))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose full service with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, resources...)

			tokenOAUTH, err := getOAUTHToken(t, oauth2Cfg)
			require.NoError(err)
			require.NotNil(tokenOAUTH)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", tokenOAUTH.AccessToken), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)
			batch.DeleteResources(k8sClient, resources...)

			assert.NoError(tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose service with OAUTH and JWT on speficic paths", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			oauthStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "jwt-oauth", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, oauthStrategyApiruleResource...)

			tokenOAUTH, err := getOAUTHToken(t, oauth2Cfg)
			require.NoError(err)
			require.NotNil(tokenOAUTH)

			tokenJWT, err := jwt.Authenticate(jwtConfig.IdProviderConfig)
			if err != nil {
				t.Fatalf("failed to fetch and id_token. %s", err.Error())
			}

			assert.Nil(err)
			assert.NotNil(tokenJWT)

			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/headers", testID, conf.Domain), fmt.Sprintf("Bearer %s", tokenOAUTH.AccessToken), defaultHeaderName))
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/image", testID, conf.Domain), fmt.Sprintf("Bearer %s", tokenJWT), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)

		})

		t.Run("Expose service with OAUTH and JWT on the same path", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			oauthStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthOnePathApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "jwt-oauth-one-path", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, oauthStrategyApiruleResource...)
			tokenOAUTH, err := getOAUTHToken(t, oauth2Cfg)
			require.NoError(err)
			require.NotNil(tokenOAUTH)

			tokenJWT, err := jwt.Authenticate(jwtConfig.IdProviderConfig)
			if err != nil {
				t.Fatalf("failed to fetch and id_token. %s", err.Error())
			}

			assert.Nil(err)
			assert.NotNil(tokenJWT)

			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/image", testID, conf.Domain), tokenOAUTH.AccessToken, "oauth2-access-token"))
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/image", testID, conf.Domain), fmt.Sprintf("Bearer %s", tokenJWT), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose service with OAUTH and update to plain access ", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, resources...)
			token, err := getOAUTHToken(t, oauth2Cfg)

			require.NoError(err)
			require.NotNil(token)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", token.AccessToken), defaultHeaderName))

			//Update API to give plain access
			namePrefix := strings.TrimSuffix(resources[0].GetName(), "-"+testID)

			unsecuredApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, unsecuredApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			batch.DeleteResources(k8sClient, commonResources...)
			batch.DeleteResources(k8sClient, unsecuredApiruleResource...)

			assert.NoError(tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, securedApiruleResource...)

			token, err := getOAUTHToken(t, oauth2Cfg)
			require.NoError(err)
			require.NotNil(token)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", token.AccessToken), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 and JWT strategy on paths", func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				Namespace        string
				NamePrefix       string
				TestID           string
				Domain           string
				GatewayName      string
				GatewayNamespace string
			}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain, GatewayName: conf.GatewayName,
				GatewayNamespace: conf.GatewayNamespace})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, securedApiruleResource...)

			oauth, err := getOAUTHToken(t, oauth2Cfg)
			require.NoError(err)
			require.NotNil(oauth)

			tokenJWT, err := jwt.Authenticate(jwtConfig.IdProviderConfig)
			if err != nil {
				t.Fatalf("failed to fetch and id_token. %s", err.Error())
			}

			assert.Nil(err)
			assert.NotNil(tokenJWT)

			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/headers", testID, conf.Domain), fmt.Sprintf("Bearer %s", oauth.AccessToken), defaultHeaderName))
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s/image", testID, conf.Domain), fmt.Sprintf("Bearer %s", tokenJWT), defaultHeaderName))
			batch.DeleteResources(k8sClient, commonResources...)
		})

	})
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

func getOAUTHToken(t *testing.T, oauth2Cfg clientcredentials.Config) (*oauth2.Token, error) {
	var tokenOAUTH *oauth2.Token
	err := retry.Do(
		func() error {
			token, err := oauth2Cfg.Token(context.Background())
			if err != nil {
				t.Errorf("Error during Token retrival: %+v", err)
				return err
			}
			tokenOAUTH = token
			return nil
		},
		retry.Delay(500*time.Millisecond), retry.Attempts(3))
	return tokenOAUTH, err
}

func lookupIPOfTestPod() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	var ips []net.IP
	ips, err = net.LookupIP(hostname)
	if err != nil {
		return "", err
	}

	if len(ips) != 1 {
		return "", errors.New("not exactly 1 ip address found for host")
	}

	return ips[0].String(), nil
}
