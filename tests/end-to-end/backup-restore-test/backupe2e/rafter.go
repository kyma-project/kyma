package backupe2e

import (
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type rafterTest struct {
	client dynamic.Interface
}

var (
	_ BackupTest = NewRafterTest(nil)

	rafterWaitTimeout = 4 * time.Minute

	rafterAssetGroupName = "backup-asset-group"
	rafterAssetName      = "backup-asset"
	rafterBucketName     = "backup-bucket"

	rafterSourceURL = "http://rafter-controller-manager.kyma-system.svc.cluster.local:8080/metrics"
)

func NewRafterTest(client dynamic.Interface) *rafterTest {
	return &rafterTest{client: client}
}

func (rt *rafterTest) CreateResources(namespace string) {
	So(rt.CreateResourcesError(namespace), ShouldBeNil)
}

func (rt *rafterTest) CreateResourcesError(namespace string) error {
	// No support for cluster-wide resources in backup/restore testing framework
	for _, fn := range []func(string) error{
		rt.createAssetGroup,
		rt.createBucket,
		rt.createAsset,
	} {
		err := fn(namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *rafterTest) TestResources(namespace string) {
	So(rt.TestResourcesError(namespace), ShouldBeNil)
}

func (rt *rafterTest) TestResourcesError(namespace string) error {
	for _, fn := range []func(string) error{
		rt.testAssetGroup,
		rt.testBucket,
		rt.testAsset,
	} {
		err := fn(namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rt *rafterTest) createAssetGroup(namespace string) error {
	assetGroup := &v1beta1.AssetGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rafterAssetGroupName,
			Namespace: namespace,
		},
		Spec: v1beta1.AssetGroupSpec{
			CommonAssetGroupSpec: v1beta1.CommonAssetGroupSpec{
				DisplayName: "Test Asset Group",
				Description: "Please, backup me, please!",
				Sources: []v1beta1.Source{
					{
						Type: "metrics",
						Name: "metrics",
						Mode: v1beta1.AssetGroupSingle,
						URL:  rafterSourceURL,
					},
				},
			},
		},
	}

	return rt.create(v1beta1.GroupVersion.WithResource("assetgroup"), namespace, assetGroup)
}

func (rt *rafterTest) createBucket(namespace string) error {
	bucket := &v1beta1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rafterBucketName,
			Namespace: namespace,
		},
		Spec: v1beta1.BucketSpec{
			CommonBucketSpec: v1beta1.CommonBucketSpec{
				Region: v1beta1.BucketRegionEUCentral1,
				Policy: v1beta1.BucketPolicyReadOnly,
			},
		},
	}

	return rt.create(v1beta1.GroupVersion.WithResource("bucket"), namespace, bucket)
}

func (rt *rafterTest) createAsset(namespace string) error {
	asset := &v1beta1.Asset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rafterAssetName,
			Namespace: namespace,
		},
		Spec: v1beta1.AssetSpec{
			CommonAssetSpec: v1beta1.CommonAssetSpec{
				Source: v1beta1.AssetSource{
					Mode: v1beta1.AssetSingle,
					URL:  rafterSourceURL,
				},
				BucketRef: v1beta1.AssetBucketRef{
					Name: rafterBucketName,
				},
			},
		},
	}

	return rt.create(v1beta1.GroupVersion.WithResource("asset"), namespace, asset)
}

func (rt *rafterTest) testAssetGroup(namespace string) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		assetGroup := &v1beta1.AssetGroup{}
		err := rt.get(v1beta1.GroupVersion.WithResource("assetgroup"), namespace, rafterAssetGroupName, assetGroup)
		if err != nil {
			return false, err
		}

		if assetGroup.Status.Phase != v1beta1.AssetGroupReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		return errors.Wrapf(err, "while waiting for ready AssetGroup resource")
	}

	return nil
}

func (rt *rafterTest) testBucket(namespace string) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		bucket := &v1beta1.Bucket{}
		err := rt.get(v1beta1.GroupVersion.WithResource("bucket"), namespace, rafterBucketName, bucket)
		if err != nil {
			return false, err
		}

		if bucket.Status.Phase != v1beta1.BucketReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Bucket resource")
	}

	return nil
}

func (rt *rafterTest) testAsset(namespace string) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		asset := &v1beta1.Asset{}
		err := rt.get(v1beta1.GroupVersion.WithResource("asset"), namespace, rafterAssetGroupName, asset)
		if err != nil {
			return false, err
		}

		if asset.Status.Phase != v1beta1.AssetReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Asset resource")
	}

	return nil
}

func (rt *rafterTest) create(resource schema.GroupVersionResource, namespace string, obj runtime.Object) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return errors.Wrapf(err, "while converting %s to unstructured", obj.GetObjectKind().GroupVersionKind().String())
	}

	_, err = rt.client.Resource(resource).Namespace(namespace).Create(&unstructured.Unstructured{Object: u}, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "while creating resource in namespace %s", namespace)
	}

	return nil
}

func (rt *rafterTest) get(resource schema.GroupVersionResource, namespace, name string, obj runtime.Object) error {
	uns, err := rt.client.Resource(resource).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while accessing %s/%s", namespace, name)
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(uns.Object, obj)
	if err != nil {
		return errors.Wrapf(err, "while converting unstructured to %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	return nil
}
