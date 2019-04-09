package testsuite

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	BucketName        = "e2ebackup-bucket"
	ClusterBucketName = "e2ebackup-cluster-bucket"
	CommonAssetPrefix = "e2ebackup-asset"
	WaitTimeout       = 3 * time.Minute
)

type TestSuite struct {
	bucket        *bucket
	clusterBucket *clusterBucket
	asset         *asset
	clusterAsset  *clusterAsset
	assetDetails  assetData
}

func New(restConfig *rest.Config, namespace string, t *testing.T) (*TestSuite, error) {
	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	b := newBucket(dynamicCli, BucketName, namespace, WaitTimeout)
	cb := newClusterBucket(dynamicCli, ClusterBucketName, WaitTimeout)
	a := newAsset(dynamicCli, BucketName, namespace, WaitTimeout)
	ca := newClusterAsset(dynamicCli, ClusterBucketName, WaitTimeout)

	return &TestSuite{
		bucket:        b,
		clusterBucket: cb,
		asset:         a,
		clusterAsset:  ca,
	}, nil
}

func (t *TestSuite) CreateBuckets() error {
	err := t.bucket.Create()
	if err != nil {
		return err
	}

	err = t.clusterBucket.Create()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForBucketsReady() error {
	err := t.bucket.WaitForStatusReady()
	if err != nil {
		return err
	}

	err = t.clusterBucket.WaitForStatusReady()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) DeleteClusterBucket() error {
	err := t.clusterBucket.Delete()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForClusterBucketDeleted() error {
	err := t.clusterBucket.WaitForDeleted()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) CreateAssets() error {
	t.assetDetails = fixSimpleAssetData(CommonAssetPrefix)

	err := t.asset.Create(t.assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.Create(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForAssetsReady() error {
	err := t.asset.WaitForStatusReady(t.assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.WaitForStatusReady(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) DeleteClusterAsset() error {
	err := t.clusterAsset.Delete(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForClusterAssetDeleted() error {
	err := t.clusterAsset.WaitForDeleted(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}
