package asset_store

import (
	"k8s.io/client-go/dynamic"
	"github.com/sirupsen/logrus"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"fmt"
	"time"
)

const (
	DocsTopicName        = "e2ebackup-docs-topic"
	ClusterDocsTopicName = "e2ebackup-cluster-docs-topic"
	WaitTimeout          = 4 * time.Minute
)

type AssetStoreUpgradeTest struct {
	dynamicInterface dynamic.Interface
}

type assetStoreFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	bucket   		*docsTopic
	clusterBucket 	*clusterDocsTopic
	asset   		*docsTopic
	clusterAsset 	*clusterDocsTopic
}

func NewAssetStoreUpgradeTest(dynamicCli dynamic.Interface) *AssetStoreUpgradeTest {
	return &AssetStoreUpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

func (ut *AssetStoreUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).createResources()
}

func (ut *AssetStoreUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}

func (ut *AssetStoreUpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *assetStoreFlow {
	return &assetStoreFlow{
		namespace: namespace,
		log: log,
		stop: stop,
		bucket: newClusterDocsTopicClient(ut.dynamicInterface),
		clusterBucket: newDocsTopic(ut.dynamicInterface, namespace),
		asset:
			clus
	}
}

func (f *assetStoreFlow) createResources() error {
	for _, t := range []struct{
		log string
		fn func(spec v1alpha1.CommonDocsTopicSpec) error
	}{
		{
			log: fmt.Sprintf("Creating ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn: f.clusterDocsTopic.create,
		},
		{
			log: fmt.Sprintf("Creating DocsTopic %s", f.docsTopic.name),
			fn: f.clusterDocsTopic.create,
		},
	}{
		f.log.Infof(t.log)
		err := t.fn(commonDocsTopicSpec)
		if err != nil {
			return err
		}
	}

	return nil
}