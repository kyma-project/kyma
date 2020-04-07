package ory

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
)

const testApiRuleFile = "test-apirule.yaml"

//TODO: Move it to a separate package (e.g: ../apigateway) once dependencies on code in ./pkg can be handled properly in this project.
//For now it's in the `ory` package because of these shared dependencies.
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

func (agt *ApiGateway) CreateResources(t *testing.T, namespace string) {
	agt.newCreateScenario(t, namespace).run()
}

func (agt *ApiGateway) TestResources(t *testing.T, namespace string) {
	agt.newTestScenario(t, namespace).run()
}

func (agt *ApiGateway) newCreateScenario(t *testing.T, namespace string) phaseRunner {

	brs := newBackupRestoreScenario(agt.k8sClient, agt.batch, agt.commonRetryOpts, namespace, "apigateway", "create")
	sc := &apiGatewayScenario{brs}
	return phaseRunner{sc.runFunc(t, sc.createResources())}
}

func (agt *ApiGateway) newTestScenario(t *testing.T, namespace string) phaseRunner {

	brs := newBackupRestoreScenario(agt.k8sClient, agt.batch, agt.commonRetryOpts, namespace, "apigateway", "test")
	sc := &apiGatewayScenario{brs}
	return phaseRunner{sc.runFunc(t, sc.testResources())}
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
		ags.verifyTestAppNoAccess,
		ags.verifyTestAppSecuredAccess,
	}
}

func (ags *apiGatewayScenario) createTestApiRule(t *testing.T) error {
	ags.log("Creating ApiRule for accessing test Application with an Access Token")
	testApiRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		t, testApiRuleFile, ags.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName, Domain string }{TestNamespace: ags.config.testNamespace, TestAppName: ags.config.testAppName, Domain: ags.getDomain(t)})

	if err != nil {
		return err
	}

	ags.clients.batch.CreateResources(ags.clients.k8sClient, testApiRuleResource...)

	return nil
}

func (ags *apiGatewayScenario) verifyTestAppNoAccess(t *testing.T) error {

	ags.log("Calling test application via external Virtual Service URL with invalid Access Token")
	testAppURL := ags.getSecuredTestAppURL(t)
	ags.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 401
	var accessToken = "Bearer Invalid"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	return ags.callWithClient(t, client, testAppURL, expectedStatusCode, accessToken)
}

func (ags *apiGatewayScenario) verifyTestAppSecuredAccess(t *testing.T) error {

	ags.log("Calling test application via external Virtual Service URL with Acces Token")
	testAppURL := ags.getSecuredTestAppURL(t)
	ags.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200
	var accessToken = fmt.Sprintf("Bearer %s", ags.data.accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	return ags.callWithClient(t, client, testAppURL, expectedStatusCode, accessToken)
}

func (ags *apiGatewayScenario) getSecuredTestAppURL(t *testing.T) string {
	return fmt.Sprintf("https://%s.%s/headers", ags.config.testAppName, ags.getDomain(t))
}

func (ags *apiGatewayScenario) getDomain(t *testing.T) string {
	domain := os.Getenv("DOMAIN")
	require.NotEmpty(t, domain)
	return domain
}
