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

type assetGroup struct {
	resCli      *resource.Resource
	BucketName  string
	Name        string
	Namespace   string
	waitTimeout time.Duration
}

func newAssetGroup(dynamicCli dynamic.Interface, name, namespace string, bucketName string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *assetGroup {
	return &assetGroup{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assetgroups",
		}, namespace, logFn),
		waitTimeout: waitTimeout,
		BucketName:  bucketName,
		Namespace:   namespace,
		Name:        name,
	}
}

func (ag *assetGroup) Create(assets []assetData, testID string, callbacks ...func(...interface{})) (string, error) {
	assetSources := make([]v1beta1.Source, 0)
	for _, asset := range assets {
		assetSources = append(assetSources, v1beta1.Source{
			Name: v1beta1.AssetGroupSourceName(asset.Name),
			URL:  asset.URL,
			Mode: v1beta1.AssetGroupSourceMode(asset.Mode),
			Type: asset.Type,
		})
	}

	assetGr := &v1beta1.AssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ag.Name,
			Namespace: ag.Namespace,
			Labels: map[string]string{
				"test-id": testID,
			},
		},
		Spec: v1beta1.AssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				BucketRef: v1beta1.AssetGroupBucketRef{
					Name: ag.BucketName,
				},
				DisplayName: "Test Asset Group",
				Description: "Gettin' ready for you!!",
				Sources:     assetSources,
			},
		},
	}

	resourceVersion, err := ag.resCli.Create(assetGr, callbacks...)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating AssetGroup %s in namespace %s", ag.Name, ag.Namespace)
	}

	return resourceVersion, nil
}

func (ag *assetGroup) WaitForStatusReady(initialResourceVersion string, callbacks ...func(...interface{})) error {
	waitForStatusReady := buildWaitForStatusesReady(ag.resCli.ResCli, ag.waitTimeout, ag.Name)
	err := waitForStatusReady(initialResourceVersion, callbacks...)
	return err
}

func (ag *assetGroup) DeleteLeftovers(testId string, callbacks ...func(...interface{})) error {
	deleteLeftovers := buildDeleteLeftovers(ag.resCli.ResCli, ag.waitTimeout)
	err := deleteLeftovers(testId, callbacks...)
	return err
}

func (ag *assetGroup) Get() (*v1beta1.AssetGroup, error) {
	u, err := ag.resCli.Get(ag.Name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.AssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, err
		}

		return nil, errors.Wrapf(err, "while converting AssetGroup %s", ag.Name)
	}

	return &res, nil
}

func (ag *assetGroup) Delete(callbacks ...func(...interface{})) error {
	err := ag.resCli.Delete(ag.Name, ag.waitTimeout, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while deleting AssetGroup %s in namespace %s", ag.Name, ag.Namespace)
	}

	return nil
}
