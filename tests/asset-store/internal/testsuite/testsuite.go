package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/namespace"
	"github.com/kyma-project/kyma/tests/asset-store/pkg/upload"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"strings"
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

	a := newAsset(dynamicCli, cfg.Namespace)
	ca := newClusterAsset(dynamicCli)

	return &TestSuite{
		namespace:ns,
		bucket:b,
		clusterBucket:cb,
		fileUpload:newFileUpload(cfg.UploadServiceUrl),
		asset:a,
		clusterAsset:ca,

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

	// Upload test files with asset-upload-service
	uploadResult, err := t.fileUpload.Do()
	if err != nil {
		return err
	}

	assetDetails := t.convertToAssetDetails(uploadResult)

	err = t.asset.Create(assetDetails)
	if err != nil {
		return err
	}

	err = t.clusterAsset.Create(assetDetails)
	if err != nil {
		return err
	}

	// TODO: Validate no errros files.Errors

	// Create clusterAsset CR (single file and package)

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

func (t *TestSuite) convertToAssetDetails(response *upload.Response) []assetDetails {
	var assets []assetDetails
	for _, file := range response.UploadedFiles {
		var mode v1alpha2.AssetMode
		if strings.HasSuffix(file.FileName, ".tar.gz") {
			mode = v1alpha2.AssetPackage
		} else {
			mode = v1alpha2.AssetSingle
		}

		asset := assetDetails{
			Name: file.FileName,
			URL: file.RemotePath,
			Mode:mode,
			Bucket:t.bucket.name,
		}
		assets = append(assets, asset)
	}

	return assets
}
