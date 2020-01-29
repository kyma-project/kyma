package ory

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	. "github.com/smartystreets/goconvey/convey"
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

func (osd *oryScenario) createResources() []scenarioStep {

	return []scenarioStep{
		osd.createTestApp,
		osd.createTestAppRule,
		osd.registerOAuth2Client,
	}
}

func (osd *oryScenario) testResources() []scenarioStep {
	return []scenarioStep{
		osd.readOAuth2ClientData,
		osd.fetchAccessToken,
		osd.verifyTestAppDirectAccess,
		osd.verifyTestAppSecuredAccess,
	}
}

func (osd *oryScenario) createTestAppRule() error {
	osd.log("Creating Oathkeeper rule for accessing test Application with an Access Token")
	testAppRuleResource, err := manifestprocessor.ParseFromFileWithTemplate(
		testAppRuleFile, osd.config.manifestsDirectory, resourceSeparator,
		struct{ TestNamespace, TestAppName string }{TestNamespace: osd.config.testNamespace, TestAppName: osd.config.testAppName})

	if err != nil {
		return err
	}

	osd.clients.batch.CreateResources(osd.clients.k8sClient, testAppRuleResource...)

	return nil
}

func (osd *oryScenario) verifyTestAppSecuredAccess() error {

	osd.log("Calling test application via Oathkeeper with Acces Token")
	testAppURL := osd.getSecuredTestAppURL()

	osd.log(fmt.Sprintf("Secured testApp URL (via oathkeeper): %s", testAppURL))
	const expectedStatusCode = 200

	client := &http.Client{}
	req, err := http.NewRequest("GET", testAppURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", osd.data.accessToken))

	resp, err := osd.retryHttpCall(func() (*http.Response, error) {
		return client.Do(req)
	}, expectedStatusCode)
	So(err, ShouldBeNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	So(err, ShouldBeNil)
	osd.log(fmt.Sprintf("Response from /headers endpoint:\n%s", string(body)))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (osd *oryScenario) getSecuredTestAppURL() string {
	return fmt.Sprintf("http://%s/ory-backup-tests-rule", oathkeeperProxyServiceName)
}
