package testsuite

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/namespace"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Config struct {
	Namespace string `envconfig:"default=test-asset-store"`
	BucketName string `envconfig:"default=test-bucket"`
	AssetName string `envconfig:"default=test-asset"`
	ClusterBucketName string `envconfig:"default=test-cluster-bucket"`
	ClusterAssetName string `envconfig:"default=test-cluster-asset"`
	UploadServiceUrl string `envconfig:"default=http://localhost:3000/v1/upload"`
}

type TestSuite struct {
	namespace *namespace.Namespace
	bucket *bucket
	clusterBucket *clusterBucket
	fileUpload *fileUpload
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
	b := newBucket(dynamicCli, cfg.BucketName, cfg.Namespace)
	cb := newClusterBucket(dynamicCli, cfg.ClusterBucketName)

	a := newAsset(dynamicCli, cfg.Namespace)
	ca := newClusterAsset(dynamicCli)

	return &TestSuite{
		namespace:ns,
		bucket:b,
		clusterBucket:cb,
		fileUpload:newFileUpload(cfg.UploadServiceUrl),
		asset:a,
		clusterAsset:ca,

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

	err = t.createAssets()
	if err != nil {
		return err
	}

	// Check if assets have been uploaded

	err = t.deleteAssets()
	if err != nil {
		return err
	}


	// See if files are gone

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

	// Namespace
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
	uploadResult, err := t.fileUpload.Do()
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

	err = t.clusterAsset.Create(t.assetDetails)
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

	err = t.clusterAsset.Delete(t.assetDetails)
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

