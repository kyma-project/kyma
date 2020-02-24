package ory

import (
	"fmt"
	"log"
	"net/http"
	"testing"

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

func (hct *HydraOathkeeper) CreateResources(t *testing.T, namespace string) {
	hct.newCreateScenario(t, namespace).run()
}

func (hct *HydraOathkeeper) TestResources(t *testing.T, namespace string) {
	hct.newTestScenario(t, namespace).run()
}

func (hct *HydraOathkeeper) newCreateScenario(t *testing.T, namespace string) phaseRunner {
	brs := newBackupRestoreScenario(hct.k8sClient, hct.batch, hct.commonRetryOpts, namespace, "ory", "create")
	sc := &oryScenario{brs}
	return phaseRunner{sc.runFunc(t, sc.createResources())}
}

func (hct *HydraOathkeeper) newTestScenario(t *testing.T, namespace string) phaseRunner {
	brs := newBackupRestoreScenario(hct.k8sClient, hct.batch, hct.commonRetryOpts, namespace, "ory", "test")
	sc := &oryScenario{brs}
	return phaseRunner{sc.runFunc(t, sc.testResources())}
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

func (osc *oryScenario) createTestAppRule(t *testing.T) error {
	osc.log("Creating Oathkeeper rule for accessing test Application with an Access Token")
	testAppRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		t, testAppRuleFile, osc.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: osc.config.testNamespace, TestAppName: osc.config.testAppName})

	if err != nil {
		return err
	}

	osc.clients.batch.CreateResources(osc.clients.k8sClient, testAppRuleResource...)

	return nil
}

func (osc *oryScenario) verifyTestAppSecuredAccess(t *testing.T) error {

	osc.log("Calling test application via Oathkeeper with Acces Token")
	testAppURL := osc.getSecuredTestAppURL()
	osc.log(fmt.Sprintf("Test application URL: %s", testAppURL))

	const expectedStatusCode = 200
	var accessToken = fmt.Sprintf("Bearer %s", osc.data.accessToken)

	client := &http.Client{}
	return osc.callWithClient(t, client, testAppURL, expectedStatusCode, accessToken)
}

func (osc *oryScenario) getSecuredTestAppURL() string {
	return fmt.Sprintf("http://%s/ory-backup-tests-rule", oathkeeperProxyServiceName)
}
