package rafter

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"time"
)

const (
	assetGroupName        = "e2eupgrade-asset-group"
	clusterAssetGroupName = "e2eupgrade-cluster-asset-group"
	bucketName        = "e2eupgrade-bucket"
	clusterBucketName = "e2eupgrade-cluster-bucket"
	waitTimeout       = 4 * time.Minute
)

// UpgradeTest tests the Rafter resources after Kyma upgrade phase
type UpgradeTest struct {
	dynamicInterface dynamic.Interface
}

type rafterFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	bucket        *bucket
	clusterBucket *clusterBucket
	asset         *asset
	clusterAsset  *clusterAsset
	assetGroup         *assetGroup
	clusterAssetGroup  *clusterAssetGroup
}

// NewRafterUpgradeTest returns new instance of the UpgradeTest
func NewRafterUpgradeTest(dynamicCli dynamic.Interface) *UpgradeTest {
	return &UpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).createResources()
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).testResources()
}

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *rafterFlow {
	return &rafterFlow{
		namespace:     namespace,
		log:           log,
		stop:          stop,
		bucket:        newBucket(ut.dynamicInterface, namespace),
		clusterBucket: newClusterBucket(ut.dynamicInterface),
		asset:         newAsset(ut.dynamicInterface, namespace),
		clusterAsset:  newClusterAsset(ut.dynamicInterface),
		assetGroup:        newAssetGroup(ut.dynamicInterface, namespace),
		clusterAssetGroup: newClusterAssetGroup(ut.dynamicInterface),
	}
}

func (f *rafterFlow) createResources() error {
	// method doesn't create resources, because they were be created in asset-store and cms domains before upgrade - purpose for test migration job
	// below statement (return) will be removed after integration Rafter in Kyma
	return nil

	for _, t := range []struct {
		log string
		fn  func() error
	}{
		{
			log: fmt.Sprintf("Creating ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.create,
		},
		{
			log: fmt.Sprintf("Creating ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.create,
		},
		{
			log: fmt.Sprintf("Creating ClusterAssetGroup %s", f.clusterAssetGroup.name),
			fn:  f.clusterAsset.create,
		},
		{
			log: fmt.Sprintf("Creating Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.create,
		},
		{
			log: fmt.Sprintf("Creating Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.create,
		},
		{
			log: fmt.Sprintf("Creating AssetGroup %s in namespace %s", f.assetGroup.name, f.namespace),
			fn:  f.bucket.create,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *rafterFlow) testResources() error {
	for _, t := range []struct {
		log string
		fn  func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterAssetGroup %s", f.clusterAssetGroup.name),
			fn:  f.clusterAssetGroup.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of AssetGroup %s in namespace %s", f.assetGroup.name, f.namespace),
			fn:  f.assetGroup.waitForStatusReady,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn(f.stop)
		if err != nil {
			return err
		}
	}

	for _, t := range []struct {
		log string
		fn  func() error
	}{
		{
			log: fmt.Sprintf("Deleting ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.delete,
		},
		{
			log: fmt.Sprintf("Deleting ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.delete,
		},
		{
			log: fmt.Sprintf("Deleting ClusterAssetGroup %s in namespace %s", f.clusterAssetGroup.name, f.namespace),
			fn:  f.clusterAssetGroup.delete,
		},
		{
			log: fmt.Sprintf("Deleting Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.delete,
		},
		{
			log: fmt.Sprintf("Deleting Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.delete,
		},
		{
			log: fmt.Sprintf("Deleting AssetGroup %s in namespace %s", f.assetGroup.name, f.namespace),
			fn:  f.assetGroup.delete,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn()
		if err != nil {
			return err
		}
	}

	for _, t := range []struct {
		log string
		fn  func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for remove ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove ClusterAssetGroup %s", f.clusterAssetGroup.name),
			fn:  f.clusterAssetGroup.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove AssetGroup %s in namespace %s", f.assetGroup.name, f.namespace),
			fn:  f.assetGroup.waitForRemove,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn(f.stop)
		if err != nil {
			return err
		}
	}

	return nil
}
