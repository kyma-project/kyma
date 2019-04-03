package assetstore

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

//
const (
	bucketName        = "e2eupgrade-bucket"
	clusterBucketName = "e2eupgrade-cluster-bucket"
	waitTimeout       = 4 * time.Minute
)

// UpgradeTest tests the AssetStore business logic after Kyma upgrade phase
type UpgradeTest struct {
	dynamicInterface dynamic.Interface
}

type assetStoreFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	bucket        *bucket
	clusterBucket *clusterBucket
	asset         *asset
	clusterAsset  *clusterAsset
}

// NewAssetStoreUpgradeTest returns new instance of the UpgradeTest
func NewAssetStoreUpgradeTest(dynamicCli dynamic.Interface) *UpgradeTest {
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

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *assetStoreFlow {
	return &assetStoreFlow{
		namespace:     namespace,
		log:           log,
		stop:          stop,
		bucket:        newBucket(ut.dynamicInterface, namespace),
		clusterBucket: newClusterBucket(ut.dynamicInterface),
		asset:         newAsset(ut.dynamicInterface, namespace),
		clusterAsset:  newClusterAsset(ut.dynamicInterface),
	}
}

func (f *assetStoreFlow) createResources() error {
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
			log: fmt.Sprintf("Creating Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.create,
		},
		{
			log: fmt.Sprintf("Creating Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.create,
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

func (f *assetStoreFlow) testResources() error {
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
			log: fmt.Sprintf("Waiting for Ready status of Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.waitForStatusReady,
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
			log: fmt.Sprintf("Deleting Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.delete,
		},
		{
			log: fmt.Sprintf("Deleting ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.delete,
		},
		{
			log: fmt.Sprintf("Deleting Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.delete,
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
			log: fmt.Sprintf("Waiting for remove Bucket %s in namespace %s", f.bucket.name, f.namespace),
			fn:  f.bucket.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove Asset %s in namespace %s", f.asset.name, f.namespace),
			fn:  f.asset.waitForRemove,
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
