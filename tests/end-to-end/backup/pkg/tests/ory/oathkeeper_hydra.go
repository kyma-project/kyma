package ory

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/client"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/manifestprocessor"
	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory/pkg/resource"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/dynamic"
)

const testAppRuleFile = "test-rule.yaml"
const oathkeeperProxyServiceName = "ory-oathkeeper-proxy.kyma-system.svc.cluster.local:4455"

type HydraOathkeeper struct {
	k8sClient       dynamic.Interface
	batch           *resource.Batch
	commonRetryOpts []retry.Option
}

type oryScenario struct {
	backupRestoreScenario
}

func NewHydraOathkeeperTest() (*HydraOathkeeper, error) {
	log.Println("Starting ORY Oathkeeper/Hydra test")

	commonRetryOpts := getCommonRetryOpts()
	resourceManager := &resource.Manager{RetryOptions: commonRetryOpts}
	batch := &resource.Batch{
		resourceManager,
	}

	return &HydraOathkeeper{client.GetDynamicClient(), batch, commonRetryOpts}, nil
}

func (hct *HydraOathkeeper) CreateResources(namespace string) {
	run(hct.newScenario(namespace).createResources())
}

func (hct *HydraOathkeeper) TestResources(namespace string) {
	run(hct.newScenario(namespace).testResources())
}

func (hct *HydraOathkeeper) newScenario(namespace string) *oryScenario {
	clients := &scenarioClients{
		k8sClient: hct.k8sClient,
		batch:     hct.batch,
	}

	config := &scenarioConfig{
		hydraURL:           getHydraURL(),
		manifestsDirectory: getManifestsDirectory(),
		commonRetryOpts:    hct.commonRetryOpts,
		testNamespace:      namespace,
		testAppName:        fmt.Sprintf("%s-ory", testAppNamePrefix),
		testSecretName:     fmt.Sprintf("%s-ory", testSecretNamePrefix),
	}
	data := &scenarioData{}

	return &oryScenario{
		backupRestoreScenario{
			clients,
			config,
			data,
		},
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
	log.Println("Creating Oathkeeper rule for accessing test Application with an Access Token")
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

	log.Println("Calling test application via Oathkeeper with Acces Token")
	testAppURL := osd.getSecuredTestAppURL()
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
	log.Printf("Response from /headers endpoint:\n%s", string(body))
	So(resp.StatusCode, ShouldEqual, expectedStatusCode)

	return nil
}

func (osd *oryScenario) getSecuredTestAppURL() string {
	securedAppURL := fmt.Sprintf("http://%s/ory-backup-tests-rule", oathkeeperProxyServiceName)
	log.Printf("Using secured testApp URL (via oathkeeper): %s", securedAppURL)
	return securedAppURL
}
