package testsuite

import (
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
}

type TestSuite struct {
	namespace *namespace.Namespace
	bucket *bucket
	clusterBucket *clusterBucket

	coreCli corev1.CoreV1Interface
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

	return &TestSuite{
		namespace:ns,
		bucket:b,
		clusterBucket:cb,

		coreCli:coreCli,
		dynamicCli:dynamicCli,
		cfg: cfg,
	}, nil
}

func (t *TestSuite) Run() error {
	err := t.namespace.Create()
	if err != nil {
		return err
	}

	err = t.bucket.Create()
	if err != nil {
		return err
	}

	err = t.clusterBucket.Create()
	if err != nil {
		return err
	}


	// Upload test data with upload service

	// Create asset CR (maybe more? single file and package)

	// Check if assets have been uploaded

	// Delete assets

	// See if files are gone

	return nil
}

func (t *TestSuite) Cleanup() error {
	err := t.bucket.Delete()
	if err != nil {
		return err
	}

	err = t.clusterBucket.Delete()
	if err != nil {
		return err
	}

	err = t.namespace.Delete()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) UploadTestData() error {
	return nil
}

func (t *TestSuite) CreateAsset() error {
	return nil
}

func (t *TestSuite) ValidateAssetUpload() error {
	return nil
}
