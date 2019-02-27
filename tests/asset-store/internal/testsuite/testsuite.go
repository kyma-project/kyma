package testsuite

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/namespace"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"time"
)

type Config struct {
	Namespace string `envconfig:"default=test-asset-store"`
	BucketName string `envconfig:"default=test-bucket"`
	AssetName string `envconfig:"default=test-asset"`
	ClusterBucketName string `envconfig:"default=test-cluster-bucket"`
	ClusterAssetName string `envconfig:"default=test-cluster-asset"`
	UploadServiceUrl string `envconfig:"default=http://localhost:3000/v1/upload"`
	WaitTimeout  time.Duration `envconfig:"default=2m"` //TODO: Change that
}

type TestSuite struct {
	namespace *namespace.Namespace
	bucket *bucket
	clusterBucket *clusterBucket
	fileUpload *testData
	asset *asset
	clusterAsset *clusterAsset

	assetDetails []assetData
	dynamicCli dynamic.Interface

	cfg Config
}

func New(restConfig *rest.Config, cfg Config) (*TestSuite, error) {
	coreCli, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	ns := namespace.New(coreCli, cfg.Namespace)

	b := newBucket(dynamicCli, cfg.BucketName, cfg.Namespace, cfg.WaitTimeout)
	cb := newClusterBucket(dynamicCli, cfg.ClusterBucketName, cfg.WaitTimeout)
	a := newAsset(dynamicCli, cfg.Namespace, cfg.WaitTimeout)
	ca := newClusterAsset(dynamicCli, cfg.WaitTimeout)

	return &TestSuite{
		namespace:     ns,
		bucket:        b,
		clusterBucket: cb,
		fileUpload:    newTestData(cfg.UploadServiceUrl),
		asset:         a,
		clusterAsset:  ca,

		dynamicCli:dynamicCli,
		cfg: cfg,
	}, nil
}

func (t *TestSuite) Run() error {
	err := t.namespace.Create()
	if err != nil {
		return err
	}

	err = t.createBuckets()
	if err != nil {
		return err
	}

	err = t.waitForBucketsReady()
	if err != nil {
		return err
	}

	err = t.createAssets()
	if err != nil {
		return err
	}

	err = t.validateFiles(true)
	if err != nil {
		return err
	}

	err = t.deleteAssets()
	if err != nil {
		return err
	}

	err = t.validateFiles(false)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) Cleanup() error {
	err := t.deleteAssets()
	if err != nil {
		return err
	}

	err = t.deleteBuckets()
	if err != nil {
		return err
	}

	err = t.namespace.Delete()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) createBuckets() error {
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

func (t *TestSuite) createAssets() error {
	uploadResult, err := t.fileUpload.Upload()
	if err != nil {
		return err
	}

	if len(uploadResult.Errors) > 0 {
		return fmt.Errorf("during file upload: %+v", uploadResult.Errors)
	}

	t.assetDetails = convertToAssetDetails(uploadResult)

	err = t.asset.Create(t.assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.CreateMany(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) waitForAssetsReady() error {
	err := t.asset.WaitForStatusReady(t.assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.WaitForStatusesReady(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) validateFiles(shouldExist bool) error {
	err := t.asset.VerifyUploadedAssets(t.assetDetails, shouldExist)
	if err != nil {
		return err
	}

	err = t.clusterAsset.VerifyUploadedAssets(t.assetDetails, shouldExist)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) waitForBucketsReady() error {
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

func (t *TestSuite) deleteAssets() error {
	err := t.asset.Delete(t.assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.DeleteMany(t.assetDetails)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) deleteBuckets() error {
	err := t.bucket.Delete()
	if err != nil {
		return err
	}

	err = t.clusterBucket.Delete()
	if err != nil {
		return err
	}

	return nil
}

