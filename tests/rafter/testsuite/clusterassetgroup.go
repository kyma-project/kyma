package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/tests/rafter/pkg/resource"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterAssetGroup struct {
	resCli            *resource.Resource
	ClusterBucketName string
	Name              string
	waitTimeout       time.Duration
}

func newClusterAssetGroup(dynamicCli dynamic.Interface, name, bucketName string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *clusterAssetGroup {
	return &clusterAssetGroup{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassetgroups",
		}, "", logFn),
		waitTimeout:       waitTimeout,
		ClusterBucketName: bucketName,
		Name:              name,
	}
}

func (cag *clusterAssetGroup) Create(assets []assetData, testID string, callbacks ...func(...interface{})) (string, error) {
	assetSources := make([]v1beta1.Source, 0)
	for _, asset := range assets {
		assetSources = append(assetSources, v1beta1.Source{
			Name: v1beta1.AssetGroupSourceName(asset.Name),
			URL:  asset.URL,
			Mode: v1beta1.AssetGroupSourceMode(asset.Mode),
			Type: v1beta1.AssetGroupSourceType(asset.Type),
		})
	}

	clusterAssetGr := &v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cag.Name,
			Labels: map[string]string{
				"test-id": testID,
			},
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				Sources:     assetSources,
				Description: "Gettin' ready for you!!",
				DisplayName: "Test Cluster Asset Group",
			},
		},
	}

	resourceVersion, err := cag.resCli.Create(clusterAssetGr)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating AssetGroup %s", cag.Name)
	}

	return resourceVersion, nil
}

func (cag *clusterAssetGroup) WaitForStatusReady(initialResourceVersion string, callbacks ...func(...interface{})) error {
	waitForStatusReady := buildWaitForStatusesReady(cag.resCli.ResCli, cag.waitTimeout, cag.Name)
	err := waitForStatusReady(initialResourceVersion, callbacks...)
	return err
}

func (cag *clusterAssetGroup) Get() (*v1beta1.ClusterAssetGroup, error) {
	u, err := cag.resCli.Get(cag.Name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while converting AssetGroup %s", cag.Name)
	}

	return &res, nil
}

func (cag *clusterAssetGroup) Delete(callbacks ...func(...interface{})) error {
	err := cag.resCli.Delete(cag.Name, cag.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting AssetGroup %s", cag.Name)
	}

	return nil
}
