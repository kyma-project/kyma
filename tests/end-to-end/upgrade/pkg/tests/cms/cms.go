package cms

import (
	"k8s.io/client-go/dynamic"
	"github.com/sirupsen/logrus"
)

type CmsUpgradeTest struct {
	dynamicInterface dynamic.Interface
}

func NewCmsUpgradeTest(dynamicCli dynamic.Interface) *CmsUpgradeTest {
	return &CmsUpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

func (ut *CmsUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}

func (ut *CmsUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}