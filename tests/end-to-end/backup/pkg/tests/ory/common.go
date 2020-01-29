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

type scenarioClients struct {
	k8sClient dynamic.Interface
	batch     *resource.Batch
}

type scenarioConfig struct {
	logPrefix          string
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

type scenarioStep func() error

func getCommonRetryOpts() []retry.Option {
	return []retry.Option{
		retry.Delay(time.Duration(commonRetryDelaySec) * time.Second),
		retry.Attempts(commonRetryTimeoutSec / commonRetryDelaySec),
		retry.DelayType(retry.FixedDelay),
	}
}

func run(steps []scenarioStep) {
	for _, fn := range steps {
		err := fn()
		if err != nil {
			log.Println(err)
		}
		So(err, ShouldBeNil)
	}
}

func (sd *backupRestoreScenario) createTestApp() error {
	log.Println("Creating test application (httpbin)")
	testAppResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppFile, sd.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: sd.config.testNamespace, TestAppName: sd.config.testAppName})

	if err != nil {
		return err
	}

	sd.clients.batch.CreateResources(sd.clients.k8sClient, testAppResource...)

	return nil
}

func (sd *backupRestoreScenario) registerOAuth2Client() error {
	log.Println("Registering OAuth2 client")
	hydraClientResource, err := manifestprocessor.ParseFromFileWithTemplate(
		hydraClientFile, sd.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, SecretName string }{TestNamespace: sd.config.testNamespace, SecretName: sd.config.testSecretName})

	if err != nil {
		return err
	}

	sd.clients.batch.CreateResources(sd.clients.k8sClient, hydraClientResource...)

	return nil
}

func (sd *backupRestoreScenario) readOAuth2ClientData() error {
	log.Println("Reading OAuth2 Client Data")
	var resource = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	var unres *unstructured.Unstructured
	var err error
	err = retry.Do(func() error {
		unres, err = sd.clients.k8sClient.Resource(resource).Namespace(sd.config.testNamespace).Get(sd.config.testSecretName, metav1.GetOptions{})
		return err
	}, sd.config.commonRetryOpts...)
	So(err, ShouldBeNil)

	data := unres.Object["data"].(map[string]interface{})

	clientID, err := valueFromSecret("client_id", data)
	So(err, ShouldBeNil)

	clientSecret, err := valueFromSecret("client_secret", data)
	So(err, ShouldBeNil)

	log.Printf("Found Client with client_id: %s", clientID)

	sd.data.oauthClientID = clientID
	sd.data.oauthClientSecret = clientSecret

	return nil
}

func (sd *backupRestoreScenario) fetchAccessToken() error {
	log.Println("Fetching OAuth2 Access Token")

	oauth2Cfg := clientcredentials.Config{
		ClientID:     sd.data.oauthClientID,
		ClientSecret: sd.data.oauthClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth2/token", sd.config.hydraURL),
		Scopes:       []string{"read"},
	}

	var token *oauth2.Token
	var err error
	err = retry.Do(func() error {
		token, err = oauth2Cfg.Token(context.Background())
		return err
	}, sd.config.commonRetryOpts...)
	So(err, ShouldBeNil)
	So(token, ShouldNotBeEmpty)

	sd.data.accessToken = token.AccessToken
	log.Printf("Access Token: %s[...]", sd.data.accessToken[:15])

	return nil
}

func (sd *backupRestoreScenario) verifyTestAppDirectAccess() error {

	log.Println("Calling test application directly to ensure it works")
	testAppURL := sd.getDirectTestAppURL()
	const expectedStatusCode = 200

	resp, err := sd.retryHttpCall(func() (*http.Response, error) {
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

func (sd *backupRestoreScenario) retryHttpCall(callerFunc func() (*http.Response, error), expectedStatusCode int) (*http.Response, error) {

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

	}, sd.config.commonRetryOpts...)

	return resp, finalErr
}

func (sd *backupRestoreScenario) getDirectTestAppURL() string {
	directAppURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:8000/headers", sd.config.testAppName, sd.config.testNamespace)
	log.Printf("Using direct testApp URL: %s", directAppURL)
	return directAppURL
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
