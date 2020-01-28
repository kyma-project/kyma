package ory

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/resource"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const commonRetryDelaySec = 6
const commonRetryTimeoutSec = 120

const hydraClientFile = "hydra-client.yaml"
const testAppFile = "test-app.yaml"
const testAppRuleFile = "test-rule.yaml"
const resourceSeparator = "---"
const secretResourceName = "api-gateway-tests-secret"
const testAppName = "httpbin-ory-backup-tests"
const hydraServiceName = "ory-hydra-public.kyma-system.svc.cluster.local:4444"
const oathkeeperProxyServiceName = "ory-oathkeeper-proxy.kyma-system.svc.cluster.local:4455"
const manifestsDirectory = "/assets/tests/ory/manifests"

type Config struct {
	hydraURL           string
	securedAppURL      string
	manifestsDirectory string
	commonRetryOpts    []retry.Option
}

type HydraOathkeeperTest struct {
	config *Config
	batch  *resource.Batch
}

type scenarioData struct {
	config            *Config
	batch             *resource.Batch
	k8sClient         dynamic.Interface
	namespace         string
	secretName        string
	oauthClientID     string
	oauthClientSecret string
	accessToken       string
}

type scenarioStep func() error

func NewHydraOathkeeperTest() (*HydraOathkeeperTest, error) {

	config := getConfig()

	resourceManager := &resource.Manager{RetryOptions: config.commonRetryOpts}
	batch := &resource.Batch{
		resourceManager,
	}

	return &HydraOathkeeperTest{config, batch}, nil
}

func (hct *HydraOathkeeperTest) CreateResources(namespace string) {
	hct.run(hct.newScenario(namespace).createResources())
}

func (hct *HydraOathkeeperTest) TestResources(namespace string) {
	hct.run(hct.newScenario(namespace).testResources())
}

func (hct *HydraOathkeeperTest) run(steps []scenarioStep) {
	for _, fn := range steps {
		err := fn()
		if err != nil {
			log.Println(err)
		}
		So(err, ShouldBeNil)
	}
}

func (hct *HydraOathkeeperTest) newScenario(namespace string) *scenarioData {
	return &scenarioData{
		config:     hct.config,
		batch:      hct.batch,
		k8sClient:  client.GetDynamicClient(),
		namespace:  namespace,
		secretName: secretResourceName,
	}
}

func (hcs *scenarioData) createResources() []scenarioStep {

	return []scenarioStep{
		hcs.createTestApp,
		hcs.createTestAppRule,
		hcs.registerOAuth2Client,
	}
}

func (hcs *scenarioData) testResources() []scenarioStep {
	return []scenarioStep{
		hcs.readOAuth2ClientData,
		hcs.fetchAccessToken,
		hcs.verifyTestAppDirectAccess,
		hcs.verifyTestAppSecuredAccess,
	}
}

func (hcs *scenarioData) createTestApp() error {
	log.Println("Creating test application (httpbin)")
	testAppResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: hcs.namespace, TestAppName: testAppName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, testAppResource...)

	return nil
}

func (hcs *scenarioData) createTestAppRule() error {
	log.Println("Creating Oathkeeper rule for accessing test Application with an Access Token")
	testAppRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppRuleFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: hcs.namespace, TestAppName: testAppName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, testAppRuleResource...)

	return nil
}

func (hcs *scenarioData) registerOAuth2Client() error {
	log.Println("Registering OAuth2 client")
	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate(
		hydraClientFile, hcs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, SecretName string }{TestNamespace: hcs.namespace, SecretName: secretResourceName})

	if err != nil {
		return err
	}

	hcs.batch.CreateResources(hcs.k8sClient, hydraClientResource...)

	return nil
}

func (hcs *scenarioData) readOAuth2ClientData() error {
	log.Println("Reading OAuth2 Client Data")
	var resource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	var unres *unstructured.Unstructured
	var err error
	err = retry.Do(func() error {
		unres, err = hcs.k8sClient.Resource(resource).Namespace(hcs.namespace).Get(hcs.secretName, metav1.GetOptions{})
		return err
	}, hcs.config.commonRetryOpts...)
	So(err, ShouldBeNil)

	data := unres.Object["data"].(map[string]interface{})

	clientID, err := valueFromSecret("client_id", data)
	So(err, ShouldBeNil)

	clientSecret, err := valueFromSecret("client_secret", data)
	So(err, ShouldBeNil)

	log.Printf("Found Client with client_id: %s", clientID)

	hcs.oauthClientID = clientID
	hcs.oauthClientSecret = clientSecret

	return nil
}

func (hcs *scenarioData) fetchAccessToken() error {
	log.Println("Fetching OAuth2 Access Token")

	oauth2Cfg := clientcredentials.Config{
		ClientID:     hcs.oauthClientID,
		ClientSecret: hcs.oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", hcs.config.hydraURL),
		Scopes:       []string{"read"},
	}

	var token *oauth2.Token
	var err error
	err = retry.Do(func() error {
		token, err = oauth2Cfg.Token(context.Background())
		return err
	}, hcs.config.commonRetryOpts...)
	So(err, ShouldBeNil)
	So(token, ShouldNotBeEmpty)

	hcs.accessToken = token.AccessToken
	log.Printf("Access Token: %s[...]", hcs.accessToken[:15])

	return nil
}

func (hcs *scenarioData) verifyTestAppDirectAccess() error {

	log.Println("Calling test application directly to ensure it works")
	testAppURL := getDirectTestAppURL(hcs.namespace)
	const expectedStatusCode = 200

	resp, err := hcs.retryHttpCall(func() (*http.Response, error) {
		return http.Get(testAppURL)
	}, expectedStatusCode)
	So(err, ShouldBeNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)
	log.Printf("Response from /headers endpoint:\n%s", string(body))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (hcs *scenarioData) verifyTestAppSecuredAccess() error {

	log.Println("Calling test application via Oathkeeper with Acces Token")
	testAppURL := hcs.config.securedAppURL
	const expectedStatusCode = 200

	client := &http.Client{}
	req, err := http.NewRequest("GET", testAppURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", hcs.accessToken))

	resp, err := hcs.retryHttpCall(func() (*http.Response, error) {
		return client.Do(req)
	}, expectedStatusCode)
	So(err, ShouldBeNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)
	log.Printf("Response from /headers endpoint:\n%s", string(body))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (hcs *scenarioData) retryHttpCall(callerFunc func() (*http.Response, error), expectedStatusCode int) (*http.Response, error) {

	var resp *http.Response
	var finalErr error

	finalErr = retry.Do(func() error {
		var err error
		resp, err = callerFunc()

		if err != nil {
			return err
		}

		if resp.StatusCode != expectedStatusCode {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				log.Printf("Error response body:\n%s", string(body))
			}
			return errors.New(fmt.Sprintf("Unexpected Status Code: %d (should be %d)", resp.StatusCode, expectedStatusCode))
		}

		return nil

	}, hcs.config.commonRetryOpts...)

	return resp, finalErr
}

func getDirectTestAppURL(testNamespace string) string {
	directAppURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:8000/headers", testAppName, testNamespace)
	log.Printf("Using direct testApp URL: %s", directAppURL)
	return directAppURL
}

func getSecuredTestAppURL() string {
	securedAppURL := fmt.Sprintf("http://%s/ory-backup-tests-rule", oathkeeperProxyServiceName)
	log.Printf("Using secured testApp URL (via oathkeeper): %s", securedAppURL)
	return securedAppURL
}

func getHydraURL() string {
	hydraURL := fmt.Sprintf("http://%s", hydraServiceName)
	log.Printf("Using Hydra URL: %s", hydraURL)
	return hydraURL
}

func getManifestsDirectory() string {
	log.Printf("Using manifest files from directory: %s", manifestsDirectory)
	return manifestsDirectory
}

func valueFromSecret(key string, dataMap map[string]interface{}) (string, error) {
	encodedValue, ok := dataMap[key].(string)
	if !ok {
		return "", errors.New("cannot read value from secret")
	}
	bres, err := base64.StdEncoding.DecodeString(encodedValue)
	return string(bres), err
}

func getConfig() *Config {

	return &Config{
		hydraURL:           getHydraURL(),
		securedAppURL:      getSecuredTestAppURL(),
		manifestsDirectory: getManifestsDirectory(),
		commonRetryOpts: []retry.Option{
			retry.Delay(time.Duration(commonRetryDelaySec) * time.Second),
			retry.Attempts(commonRetryTimeoutSec / commonRetryDelaySec),
			retry.DelayType(retry.FixedDelay),
		},
	}
}
