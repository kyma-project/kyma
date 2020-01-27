package ory

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/resource"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/oauth2/clientcredentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const hydraUrlEnvVar = "ORY_BACKUP_TEST_HYDRA_URL"
const manifestsDirEnvVar = "ORY_BACKUP_TEST_MANIFESTS_DIR"

const hydraClientFile = "hydra-client.yaml"
const testAppFile = "test-app.yaml"
const testAppRuleFile = "test-rule.yaml"
const resourceSeparator = "---"
const secretResourceName = "api-gateway-tests-secret"
const testAppName = "httpbin-ory-backup-tests"

//const manifestsDirectory = "../pkg/tests/ory/manifests"
//
//var resourceManager *resource.Manager
//var batch *resource.Batch
//var commonRetryOpts []retry.Option

type Config struct {
	hydraURL           string
	manifestsDirectory string
	commonRetryOpts    []retry.Option
}

type HydraClientTest struct {
	config *Config
	batch  *resource.Batch
}

type hydraClientScenario struct {
	config            *Config
	batch             *resource.Batch
	k8sClient         dynamic.Interface
	namespace         string
	secretName        string
	oauthClientID     string
	oauthClientSecret string
}

type scenarioStep func() error

func NewHydraClientTest() (*HydraClientTest, error) {

	config := parseExternalConfig()

	resourceManager := &resource.Manager{RetryOptions: config.commonRetryOpts}
	batch := &resource.Batch{
		resourceManager,
	}

	return &HydraClientTest{config, batch}, nil
}

func (hct *HydraClientTest) CreateResources(namespace string) {
	hct.run(hct.newScenario(namespace).createResources())
}

func (hct *HydraClientTest) TestResources(namespace string) {
	hct.run(hct.newScenario(namespace).testResources())
}

func (hct *HydraClientTest) run(steps []scenarioStep) {
	for _, fn := range steps {
		err := fn()
		if err != nil {
			log.Println(err)
		}
		So(err, ShouldBeNil)
	}
}

func (hct *HydraClientTest) newScenario(namespace string) *hydraClientScenario {
	return &hydraClientScenario{
		config:     hct.config,
		batch:      hct.batch,
		k8sClient:  client.GetDynamicClient(),
		namespace:  namespace,
		secretName: secretResourceName,
	}
}

func (hcs *hydraClientScenario) createResources() []scenarioStep {

	res := []scenarioStep{
		hcs.createTestApp,
		hcs.createTestAppRule,
		hcs.registerOAuth2Client,
	}
	res = append(res, hcs.testResources()...)
	return res
}

func (hcs *hydraClientScenario) testResources() []scenarioStep {
	return []scenarioStep{
		hcs.readOAuth2ClientData,
		hcs.fetchAccessToken,
		hcs.verifyTestAppDirectAccess,
		hcs.verifyTestAppSecuredAccess,
	}
}

func (hcs *hydraClientScenario) createTestApp() error {
	log.Println("createTestApp")
	testAppResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: hcs.namespace, TestAppName: testAppName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, testAppResource...)

	return nil
}

func (hcs *hydraClientScenario) createTestAppRule() error {
	log.Println("createTestAppRule")
	testAppRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppRuleFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: hcs.namespace, TestAppName: testAppName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, testAppRuleResource...)

	return nil
}

func (hcs *hydraClientScenario) registerOAuth2Client() error {
	log.Println("registerOAuth2Client")
	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate(
		hydraClientFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, SecretName string }{TestNamespace: hcs.namespace, SecretName: secretResourceName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, hydraClientResource...)

	return nil
}

func (hcs *hydraClientScenario) readOAuth2ClientData() error {
	log.Println("readOAuth2ClientData")
	var resource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	var unres *unstructured.Unstructured
	retry.Do(func() error {
		var err error
		unres, err = hcs.k8sClient.Resource(resource).Namespace(hcs.namespace).Get(hcs.secretName, metav1.GetOptions{})
		fmt.Println(err)
		return err
	}, hcs.config.commonRetryOpts...)

	fmt.Println("----------------------------------------!")
	data := unres.Object["data"].(map[string]interface{})
	clientID, err := valueFromSecret("client_id", data)
	if err != nil {
		return err
	}
	clientSecret, err := valueFromSecret("client_secret", data)
	if err != nil {
		return err
	}

	hcs.oauthClientID = clientID
	hcs.oauthClientSecret = clientSecret

	fmt.Println("Client ID: " + clientID)
	fmt.Println("Client Secret: " + clientSecret)

	fmt.Println("----------------------------------------")
	return nil
}

func (hcs *hydraClientScenario) fetchAccessToken() error {
	log.Println("fetchAccessToken")

	oauth2Cfg := clientcredentials.Config{
		ClientID:     hcs.oauthClientID,
		ClientSecret: hcs.oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", hcs.config.hydraURL),
		Scopes:       []string{"read"},
	}

	token, err := oauth2Cfg.Token(context.Background())
	So(err, ShouldBeNil)
	So(token, ShouldNotBeEmpty)
	log.Println("Token: " + token.AccessToken)
	return nil
}

func (hcs *hydraClientScenario) verifyTestAppDirectAccess() error {
	log.Println("verifyTestAppDirectAccess")
	return nil
}

func (hcs *hydraClientScenario) verifyTestAppSecuredAccess() error {
	log.Println("verifyTestAppSecuredAccess")
	return nil
}

func valueFromSecret(key string, dataMap map[string]interface{}) (string, error) {
	encodedValue, ok := dataMap[key].(string)
	if !ok {
		return "", errors.New("cannot read value from secret")
	}
	bres, err := base64.StdEncoding.DecodeString(encodedValue)
	return string(bres), err
}

func (hcs *hydraClientScenario) verifyToken() error {
	return nil
}

func parseExternalConfig() *Config {

	const retryDelay = 4
	const retryTimeout = 12

	hydraURL := os.Getenv(hydraUrlEnvVar)
	log.Printf("Configured with %s=%s", hydraUrlEnvVar, hydraURL)
	So(hydraURL, ShouldNotBeEmpty)

	manifestsDirectory := os.Getenv(manifestsDirEnvVar)
	log.Printf("Configured with %s=%s", manifestsDirEnvVar, manifestsDirectory)
	So(manifestsDirectory, ShouldNotBeEmpty)

	res := Config{}
	res.hydraURL = hydraURL
	res.manifestsDirectory = manifestsDirectory
	res.commonRetryOpts = []retry.Option{
		retry.Delay(time.Duration(retryDelay) * time.Second),
		retry.Attempts(retryTimeout / retryDelay),
		retry.DelayType(retry.FixedDelay),
	}

	return &res
}
