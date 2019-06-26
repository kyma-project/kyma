package backupe2eAssetStore

import (
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/asset-store/testsuite"
	restclient "k8s.io/client-go/rest"

	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
)

type assetStoreTest struct {
	restConfig *restclient.Config
	testSuite  *testsuite.TestSuite
	t          *testing.T
}

func NewAssetStoreTest(t *testing.T) (*assetStoreTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return &assetStoreTest{}, err
	}

	return &assetStoreTest{
		restConfig: restConfig,
		testSuite:  nil,
		t:          t,
	}, nil
}

func (a *assetStoreTest) CreateResources(namespace string) {
	testSuite, err := testsuite.New(a.restConfig, namespace, a.t)
	So(err, ShouldBeNil)
	a.setTestSuite(testSuite)

	err = a.testSuite.CreateBuckets()
	So(err, ShouldBeNil)

	err = a.testSuite.CreateAssets()
	So(err, ShouldBeNil)
}

func (a *assetStoreTest) TestResources(namespace string) {
	if a.testSuite == nil {
		testSuite, err := testsuite.New(a.restConfig, namespace, a.t)
		So(err, ShouldBeNil)
		a.setTestSuite(testSuite)
	}

	err := a.testSuite.WaitForBucketsReady()
	So(err, ShouldBeNil)

	err = a.testSuite.WaitForAssetsReady()
	So(err, ShouldBeNil)
}

func (a *assetStoreTest) setTestSuite(testSuite *testsuite.TestSuite) {
	a.testSuite = testSuite
}
