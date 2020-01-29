package ory

import (
	"crypto/tls"
	"fmt"
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
	agt.newCreateScenario(namespace).run()
}

func (agt *ApiGateway) TestResources(namespace string) {
	agt.newTestScenario(namespace).run()
}

func (agt *ApiGateway) newCreateScenario(namespace string) phaseRunner {

	brs := newBackupRestoreScenario(agt.k8sClient, agt.batch, agt.commonRetryOpts, namespace, "apigateway", "create")
	sc := &apiGatewayScenario{brs}
	return phaseRunner{sc.runFunc(sc.createResources())}
}

func (agt *ApiGateway) newTestScenario(namespace string) phaseRunner {

	brs := newBackupRestoreScenario(agt.k8sClient, agt.batch, agt.commonRetryOpts, namespace, "apigateway", "test")
	sc := &apiGatewayScenario{brs}
	return phaseRunner{sc.runFunc(sc.testResources())}
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
		ags.dumpOathkeeperRules,
		ags.verifyTestAppNoAccess,
		ags.verifyTestAppSecuredAccess,
	}
}

//TODO: Remove
func (ags *apiGatewayScenario) dumpOathkeeperRules() error {

	ags.log("Dumping Oathkeeper rules (DEBUG)")

	testAppURL := "http://ory-oathkeeper-api.kyma-system.svc.cluster.local:4456/rules"
	ags.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200

	client := &http.Client{}
	return ags.callWithClient(client, testAppURL, expectedStatusCode, "")

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

func (ags *apiGatewayScenario) verifyTestAppNoAccess() error {

	ags.log("Calling test application via external Virtual Service URL with invalid Access Token")
	testAppURL := ags.getSecuredTestAppURL()
	ags.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 403
	var accessToken = "Bearer Invalid"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	return ags.callWithClient(client, testAppURL, expectedStatusCode, accessToken)
}

func (ags *apiGatewayScenario) verifyTestAppSecuredAccess() error {

	ags.log("Calling test application via external Virtual Service URL with Acces Token")
	testAppURL := ags.getSecuredTestAppURL()
	ags.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200
	var accessToken = fmt.Sprintf("Bearer %s", ags.data.accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	return ags.callWithClient(client, testAppURL, expectedStatusCode, accessToken)
}

func (ags *apiGatewayScenario) getSecuredTestAppURL() string {
	return fmt.Sprintf("https://%s.%s/headers", ags.config.testAppName, ags.getDomain())
}

func (ags *apiGatewayScenario) getDomain() string {
	domain := os.Getenv("DOMAIN")
	So(domain, ShouldNotBeEmpty)
	return domain
}
