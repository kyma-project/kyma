package backupe2eAssetStore

import (
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/asset-store/testsuite"
	restclient "k8s.io/client-go/rest"

	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
)

type assetStoreTest struct {
	uuid       string
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
		uuid:       uuid.New().String(),
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
	err := a.testSuite.WaitForBucketsReady()
	So(err, ShouldBeNil)

	err = a.testSuite.WaitForAssetsReady()
	So(err, ShouldBeNil)
}

func (a *assetStoreTest) DeleteResources(namespace string) {
	err := a.testSuite.DeleteClusterBucket()
	So(err, ShouldBeNil)

	err = a.testSuite.WaitForClusterBucketDeleted()
	So(err, ShouldBeNil)

	err = a.testSuite.DeleteClusterAsset()
	So(err, ShouldBeNil)

	err = a.testSuite.WaitForClusterAssetDeleted()
	So(err, ShouldBeNil)
}

func (a *assetStoreTest) setTestSuite(testSuite *testsuite.TestSuite) {
	a.testSuite = testSuite
}
