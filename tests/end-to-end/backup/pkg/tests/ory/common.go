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

const resourceSeparator = "---"
const manifestsDirectory = "/assets/tests/ory/manifests"
const testAppFile = "test-app.yaml"
const hydraClientFile = "hydra-client.yaml"
const testAppNamePrefix = "httpbin-backup-tests"
const testSecretNamePrefix = "backup-tests-secret"
const hydraServiceName = "ory-hydra-public.kyma-system.svc.cluster.local:4444"

type scenarioStep func() error

//stepRunner is used to run a single phase of the test (create/test)
type phaseRunner struct {
	runFunc func()
}

func (sr phaseRunner) run() {
	sr.runFunc()
}

type testCommon struct {
	k8sClient       dynamic.Interface
	batch           *resource.Batch
	commonRetryOpts []retry.Option
}

type scenarioClients struct {
	k8sClient dynamic.Interface
	batch     *resource.Batch
}

type scenarioConfig struct {
	scenarioTag        string
	hydraURL           string
	manifestsDirectory string
	commonRetryOpts    []retry.Option
	testNamespace      string
	testAppName        string
	testSecretName     string
}

type scenarioData struct {
	oauthClientID     string
	oauthClientSecret string
	accessToken       string
}

type backupRestoreScenario struct {
	clients *scenarioClients
	config  *scenarioConfig
	data    *scenarioData
}

func newTestCommon() *testCommon {
	commonRetryOpts := getCommonRetryOpts()
	resourceManager := &resource.Manager{RetryOptions: commonRetryOpts}
	batch := &resource.Batch{
		resourceManager,
	}

	return &testCommon{client.GetDynamicClient(), batch, commonRetryOpts}
}

func newBackupRestoreScenario(k8sClient dynamic.Interface, batch *resource.Batch, commonRetryOpts []retry.Option, namespace, scenarioName, scenarioPhase string) *backupRestoreScenario {

	clients := &scenarioClients{
		k8sClient: k8sClient,
		batch:     batch,
	}

	config := &scenarioConfig{
		scenarioTag:        fmt.Sprintf("%s-%s", scenarioName, scenarioPhase),
		hydraURL:           getHydraURL(),
		manifestsDirectory: manifestsDirectory,
		commonRetryOpts:    commonRetryOpts,
		testNamespace:      namespace,
		testAppName:        fmt.Sprintf("%s-%s", testAppNamePrefix, scenarioName),
		testSecretName:     fmt.Sprintf("%s-%s", testSecretNamePrefix, scenarioName),
	}
	data := &scenarioData{}

	res := &backupRestoreScenario{
		clients,
		config,
		data,
	}

	res.log(fmt.Sprintf("Scenario executed in namespace: %s", namespace))
	res.log(fmt.Sprintf("Template files loaded from directory: %s", manifestsDirectory))
	return res
}

func getCommonRetryOpts() []retry.Option {
	return []retry.Option{
		retry.Delay(time.Duration(commonRetryDelaySec) * time.Second),
		retry.Attempts(commonRetryTimeoutSec / commonRetryDelaySec),
		retry.DelayType(retry.FixedDelay),
	}
}

func (brs *backupRestoreScenario) runFunc(steps []scenarioStep) func() {
	return func() {
		for _, fn := range steps {
			err := fn()
			if err != nil {
				brs.log(err)
			}
			So(err, ShouldBeNil)
		}
	}
}

func (brs *backupRestoreScenario) log(v interface{}) {
	log.Printf("[%s] %v", brs.config.scenarioTag, v)
}

func (brs *backupRestoreScenario) createTestApp() error {
	brs.log("Creating test application (httpbin)")
	testAppResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppFile, brs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: brs.config.testNamespace, TestAppName: brs.config.testAppName})

	if err != nil {
		return err
	}

	brs.clients.batch.CreateResources(brs.clients.k8sClient, testAppResource...)

	return nil
}

func (brs *backupRestoreScenario) registerOAuth2Client() error {
	brs.log("Registering OAuth2 client")
	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate(
		hydraClientFile, brs.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, SecretName string }{TestNamespace: brs.config.testNamespace, SecretName: brs.config.testSecretName})

	if err != nil {
		return err
	}

	brs.clients.batch.CreateResources(brs.clients.k8sClient, hydraClientResource...)

	return nil
}

func (brs *backupRestoreScenario) readOAuth2ClientData() error {
	brs.log("Reading OAuth2 Client Data")
	var resource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	var unres *unstructured.Unstructured
	var err error
	err = retry.Do(func() error {
		unres, err = brs.clients.k8sClient.Resource(resource).Namespace(brs.config.testNamespace).Get(brs.config.testSecretName, metav1.GetOptions{})
		if err != nil {
			brs.log(err)
		}
		return err
	}, brs.config.commonRetryOpts...)
	So(err, ShouldBeNil)

	data := unres.Object["data"].(map[string]interface{})

	clientID, err := valueFromSecret("client_id", data)
	So(err, ShouldBeNil)

	clientSecret, err := valueFromSecret("client_secret", data)
	So(err, ShouldBeNil)

	brs.log(fmt.Sprintf("Found Client with client_id: %s", clientID))

	brs.data.oauthClientID = clientID
	brs.data.oauthClientSecret = clientSecret

	return nil
}

func (brs *backupRestoreScenario) fetchAccessToken() error {
	brs.log("Fetching OAuth2 Access Token")
	tokenURL := fmt.Sprintf("%s/oauth2/token", brs.config.hydraURL)
	brs.log(fmt.Sprintf("Token URL: %s", tokenURL))

	oauth2Cfg := clientcredentials.Config{
		ClientID:     brs.data.oauthClientID,
		ClientSecret: brs.data.oauthClientSecret,
		TokenURL:     tokenURL,
		Scopes:       []string{"read"},
	}

	var token *oauth2.Token
	var err error
	err = retry.Do(func() error {
		token, err = oauth2Cfg.Token(context.Background())
		return err
	}, brs.config.commonRetryOpts...)
	So(err, ShouldBeNil)
	So(token, ShouldNotBeEmpty)

	brs.data.accessToken = token.AccessToken
	brs.log(fmt.Sprintf("Access Token: %s[...]", brs.data.accessToken[:15]))

	return nil
}

func (brs *backupRestoreScenario) verifyTestAppDirectAccess() error {

	brs.log("Calling test application directly to ensure it works")
	testAppURL := brs.getDirectTestAppURL()
	brs.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200

	client := &http.Client{}
	return brs.callWithClient(client, testAppURL, expectedStatusCode, "")
}

func (brs *backupRestoreScenario) callWithClient(client *http.Client, testAppURL string, expectedStatusCode int, accessToken string) error {

	req, err := http.NewRequest("GET", testAppURL, nil)
	if len(accessToken) > 0 {
		req.Header.Add("Authorization", accessToken)
	}

	resp, err := brs.retryHttpCall(func() (*http.Response, error) {
		return client.Do(req)
	}, expectedStatusCode)
	So(err, ShouldBeNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)
	brs.log(fmt.Sprintf("Response from endpoint:\n%s", string(body)))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (brs *backupRestoreScenario) retryHttpCall(callerFunc func() (*http.Response, error), expectedStatusCode int) (*http.Response, error) {

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
				brs.log(fmt.Sprintf("Error response body:\n%s", string(body)))
			}
			return errors.New(fmt.Sprintf("Unexpected Status Code: %d (should be %d)", resp.StatusCode, expectedStatusCode))
		}

		return nil

	}, brs.config.commonRetryOpts...)

	return resp, finalErr
}

func (brs *backupRestoreScenario) getDirectTestAppURL() string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:8000/headers", brs.config.testAppName, brs.config.testNamespace)
}

func getHydraURL() string {
	return fmt.Sprintf("http://%s", hydraServiceName)
}

func valueFromSecret(key string, dataMap map[string]interface{}) (string, error) {
	encodedValue, ok := dataMap[key].(string)
	if !ok {
		return "", errors.New("cannot read value from secret")
	}
	bres, err := base64.StdEncoding.DecodeString(encodedValue)
	return string(bres), err
}
