package ory

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	. "github.com/smartystreets/goconvey/convey"
)

const testApiRuleFile = "test-apirule.yaml"

//TODO: Move it to a separate package (e.g: apiGateway) once dependencies on code in ./pkg can be handled properly.
//For now it's in the `ory` package because of these dependencies.
type ApiGateway struct {
	*testCommon
}

type apiGatewayScenario struct {
	*backupRestoreScenario
}

func NewApiGatewayTest() (*ApiGateway, error) {
	log.Println("Starting Api-Gateway test")

	return &ApiGateway{newTestCommon()}, nil
}

func (agt *ApiGateway) CreateResources(namespace string) {
	run(agt.newScenario(namespace).createResources())
}

func (agt *ApiGateway) TestResources(namespace string) {
	run(agt.newScenario(namespace).testResources())
}

func (agt *ApiGateway) newScenario(namespace string) *apiGatewayScenario {

	brs := newBackupRestoreScenario(agt.k8sClient, agt.batch, agt.commonRetryOpts, namespace, "apigateway")
	return &apiGatewayScenario{
		brs,
	}
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
	ags.log("Creating ApiRule for accessing test Application with an Access Token")
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

	ags.log("Calling test application via Oathkeeper with Acces Token")
	testAppURL := ags.getSecuredTestAppURL()
	ags.log(fmt.Sprintf("testApp URL: %s", testAppURL))

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
	ags.log(fmt.Sprintf("Response from /headers endpoint:\n%s", string(body)))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (ags *apiGatewayScenario) getSecuredTestAppURL() string {
	return fmt.Sprintf("https://%s.%s/headers", ags.config.testAppName, ags.getDomain())
}

func (ags *apiGatewayScenario) getDomain() string {
	domain := os.Getenv("DOMAIN")
	So(domain, ShouldNotBeEmpty)
	return domain
}
