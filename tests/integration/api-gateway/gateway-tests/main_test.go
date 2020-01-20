package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
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

var resourceManager *resource.Manager

type Config struct {
	HydraAddr  string `envconfig:"TEST_HYDRA_ADDRESS"`
	User       string `envconfig:"TEST_USER_EMAIL"`
	Pwd        string `envconfig:"TEST_USER_PASSWORD"`
	ReqTimeout uint   `envconfig:"TEST_REQUEST_TIMEOUT,default=100"`
	ReqDelay   uint   `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain     string `envconfig:"DOMAIN"`
}

func TestApiGatewayIntegration(t *testing.T) {

	assert, require := assert.New(t), require.New(t)

	var conf Config
	if err := envconfig.Init(&conf); err != nil {
		t.Fatalf("Unable to setup config: %v", err)
	}

	httpClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		t.Fatalf("Unable to initialize ingress gateway client: %v", err)
	}

	oauthClientID := generateRandomString(OauthClientIDLength)
	oauthClientSecret := generateRandomString(OauthClientSecretLength)

	oauth2Cfg := clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", conf.HydraAddr),
		Scopes:       []string{"read"},
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
		OauthClientSecret string
		OauthClientID     string
	}{
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(oauthClientSecret)),
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(oauthClientID)),
	})
	if err != nil {
		panic(err)
	}

	// delete test namespace if the previous test namespace persists

	nsResourceSchema, ns, name := resource.GetResourceSchemaAndNamespace(globalCommonResources[0])
	resourceManager.DeleteResource(k8sClient, nsResourceSchema, ns, name)

	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	batch.CreateResources(k8sClient, globalCommonResources...)
	time.Sleep(time.Duration(conf.ReqDelay) * time.Second)

	hydraClientResource, err := manifestprocessor.ParseFromFile(hydraClientFile, manifestsDirectory, resourceSeparator)
	if err != nil {
		panic(err)
	}
	batch.CreateResources(k8sClient, hydraClientResource...)
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
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			//for _, commonResource := range commonResources {
			//	resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
			//	manager.UpdateResource(k8sClient, resourceSchema, ns, name, commonResource)
			//}

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose full service with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, resources...)

			token, err := oauth2Cfg.Token(context.Background())
			require.NoError(err)
			require.NotNil(token)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", token.AccessToken), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)

			assert.NoError(tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose service with OAUTH and JWT on speficic paths", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			oauthStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "jwt-oauth", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, oauthStrategyApiruleResource...)

			tokenOAUTH, err := oauth2Cfg.Token(context.Background())
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
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			oauthStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthOnePathApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "jwt-oauth-one-path", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, oauthStrategyApiruleResource...)

			tokenOAUTH, err := oauth2Cfg.Token(context.Background())
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
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, resources...)

			token, err := oauth2Cfg.Token(context.Background())
			require.NoError(err)
			require.NotNil(token)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", token.AccessToken), defaultHeaderName))

			//Update API to give plain access
			namePrefix := strings.TrimSuffix(resources[0].GetName(), "-"+testID)

			unsecuredApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, unsecuredApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			batch.DeleteResources(k8sClient, commonResources...)

			assert.NoError(tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)
			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, securedApiruleResource...)

			token, err := oauth2Cfg.Token(context.Background())
			require.NoError(err)
			require.NotNil(token)
			assert.NoError(tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), fmt.Sprintf("Bearer %s", token.AccessToken), defaultHeaderName))

			batch.DeleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 and JWT strategy on paths", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)
			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			batch.CreateResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			batch.UpdateResources(k8sClient, securedApiruleResource...)

			oauth, err := oauth2Cfg.Token(context.Background())
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
