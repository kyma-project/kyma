package ory

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
)

const testAppRuleFile = "test-rule.yaml"
const oathkeeperProxyServiceName = "ory-oathkeeper-proxy.kyma-system.svc.cluster.local:4455"

type HydraOathkeeper struct {
	*testCommon
}

type oryScenario struct {
	*backupRestoreScenario
}

func NewHydraOathkeeperTest() (*HydraOathkeeper, error) {
	log.Println("Starting ORY Oathkeeper/Hydra test")

	return &HydraOathkeeper{newTestCommon()}, nil
}

func (hct *HydraOathkeeper) CreateResources(namespace string) {
	run(hct.newScenario(namespace).createResources())
}

func (hct *HydraOathkeeper) TestResources(namespace string) {
	run(hct.newScenario(namespace).testResources())
}

func (hct *HydraOathkeeper) newScenario(namespace string) *oryScenario {
	brs := newBackupRestoreScenario(hct.k8sClient, hct.batch, hct.commonRetryOpts, namespace, "ory")
	return &oryScenario{
		brs,
	}
}

func (osc *oryScenario) createResources() []scenarioStep {

	return []scenarioStep{
		osc.createTestApp,
		osc.createTestAppRule,
		osc.registerOAuth2Client,
	}
}

func (osc *oryScenario) testResources() []scenarioStep {
	return []scenarioStep{
		osc.readOAuth2ClientData,
		osc.fetchAccessToken,
		osc.verifyTestAppDirectAccess,
		osc.verifyTestAppSecuredAccess,
	}
}

func (osc *oryScenario) createTestAppRule() error {
	osc.log("Creating Oathkeeper rule for accessing test Application with an Access Token")
	testAppRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppRuleFile, osc.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: osc.config.testNamespace, TestAppName: osc.config.testAppName})

	if err != nil {
		return err
	}

	osc.clients.batch.CreateResources(osc.clients.k8sClient, testAppRuleResource...)

	return nil
}

func (osc *oryScenario) verifyTestAppSecuredAccess() error {

	osc.log("Calling test application via Oathkeeper with Acces Token")
	testAppURL := osc.getSecuredTestAppURL()
	osc.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200
	var accessToken = fmt.Sprintf("Bearer %s", osc.data.accessToken)

	client := &http.Client{}
	return osc.callWithClient(client, testAppURL, expectedStatusCode, accessToken)
}

func (osc *oryScenario) getSecuredTestAppURL() string {
	return fmt.Sprintf("http://%s/ory-backup-tests-rule", oathkeeperProxyServiceName)
}
