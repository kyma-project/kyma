package ory

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/resource"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/dynamic"
)

const testApiRuleFile = "test-apirule.yaml"

//TODO: Move it to a separate package (e.g: apiGateway) once dependencies on code in ./pkg can be handled properly.
//For now it's in the `ory` package because of these dependencies.
type ApiGateway struct {
	k8sClient       dynamic.Interface
	batch           *resource.Batch
	commonRetryOpts []retry.Option
}

type apiGatewayScenario struct {
	backupRestoreScenario
}

func NewApiGatewayTest() (*ApiGateway, error) {
	log.Println("Starting Api-Gateway test")

	commonRetryOpts := getCommonRetryOpts()
	resourceManager := &resource.Manager{RetryOptions: commonRetryOpts}
	batch := &resource.Batch{
		resourceManager,
	}

	return &ApiGateway{client.GetDynamicClient(), batch, commonRetryOpts}, nil
}

func (agt *ApiGateway) CreateResources(namespace string) {
	run(agt.newScenario(namespace).createResources())
}

func (agt *ApiGateway) TestResources(namespace string) {
	run(agt.newScenario(namespace).testResources())
}

func (agt *ApiGateway) newScenario(namespace string) *apiGatewayScenario {

	clients := &scenarioClients{
		k8sClient: agt.k8sClient,
		batch:     agt.batch,
	}

	config := &scenarioConfig{
		hydraURL:           getHydraURL(),
		manifestsDirectory: getManifestsDirectory(),
		commonRetryOpts:    agt.commonRetryOpts,
		testNamespace:      namespace,
		testAppName:        fmt.Sprintf("%s-apigateway", testAppNamePrefix),
		testSecretName:     fmt.Sprintf("%s-apigateway", testSecretNamePrefix),
	}
	data := &scenarioData{}

	return &apiGatewayScenario{
		backupRestoreScenario{
			clients,
			config,
			data,
		},
	}
}

func (ags *apiGatewayScenario) getDomain() string {
	domain := os.Getenv("DOMAIN")
	So(domain, ShouldNotBeEmpty)
	log.Printf("Using domain value: %s", domain)
	return domain
}

func (ags *apiGatewayScenario) createResources() []scenarioStep {

	return []scenarioStep{
		ags.createTestApp,
		ags.createTestApiRule,
		ags.registerOAuth2Client,
	}
}

func (ags *apiGatewayScenario) testResources() []scenarioStep {
	return []scenarioStep{
		ags.readOAuth2ClientData,
		ags.fetchAccessToken,
		ags.verifyTestAppDirectAccess,
		ags.verifyTestAppSecuredAccess,
	}
}

func (ags *apiGatewayScenario) createTestApiRule() error {
	log.Println("Creating ApiRule for accessing test Application with an Access Token")
	testApiRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testApiRuleFile, ags.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName, Domain string }{TestNamespace: ags.config.testNamespace, TestAppName: ags.config.testAppName, Domain: ags.getDomain()})

	if err != nil {
		return err
	}

	ags.clients.batch.CreateResources(ags.clients.k8sClient, testApiRuleResource...)

	return nil
}

//TODO: Move to common.go
func (ags *apiGatewayScenario) verifyTestAppSecuredAccess() error {

	log.Println("Calling test application via Oathkeeper with Acces Token")
	testAppURL := ags.getSecuredTestAppURL()
	const expectedStatusCode = 200

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", testAppURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ags.data.accessToken))

	resp, err := ags.retryHttpCall(func() (*http.Response, error) {
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

func (ags *apiGatewayScenario) getSecuredTestAppURL() string {
	securedAppURL := fmt.Sprintf("https://%s.%s/headers", ags.config.testAppName, ags.getDomain())
	log.Printf("Using secured testApp URL (via oathkeeper): %s", securedAppURL)
	return securedAppURL
}
