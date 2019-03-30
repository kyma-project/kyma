package asset_store

import (
	"k8s.io/client-go/dynamic"
	"github.com/sirupsen/logrus"
)

type AssetStoreUpgradeTest struct {
	dynamicInterface dynamic.Interface
}

func NewAssetStoreUpgradeTest(dynamicCli dynamic.Interface) *AssetStoreUpgradeTest {
	return &AssetStoreUpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

func (ut *AssetStoreUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}

func (ut *AssetStoreUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}