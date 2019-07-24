package cms

import (
	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/backupe2e/cms/testsuite"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	. "github.com/smartystreets/goconvey/convey"
	restclient "k8s.io/client-go/rest"
)

type cmsTest struct {
	restConfig *restclient.Config
	testSuite  *testsuite.TestSuite
	t          *testing.T
}

func NewCmsTest(t *testing.T) (*cmsTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return &cmsTest{}, err
	}

	return &cmsTest{
		restConfig: restConfig,
		testSuite:  nil,
		t:          t,
	}, nil
}

func (a *cmsTest) CreateResources(namespace string) {
	testSuite, err := testsuite.New(a.restConfig, namespace, a.t)
	So(err, ShouldBeNil)
	a.setTestSuite(testSuite)

	err = a.testSuite.CreateDocsTopics()
	So(err, ShouldBeNil)
}

func (a *cmsTest) TestResources(namespace string) {
	if a.testSuite == nil {
		testSuite, err := testsuite.New(a.restConfig, namespace, a.t)
		So(err, ShouldBeNil)
		a.setTestSuite(testSuite)
	}

	err := a.testSuite.WaitForDocsTopicsReady()
	So(err, ShouldBeNil)
}

func (a *cmsTest) setTestSuite(testSuite *testsuite.TestSuite) {
	a.testSuite = testSuite
}
