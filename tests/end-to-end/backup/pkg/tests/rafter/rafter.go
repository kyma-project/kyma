package rafter

import (
	"log"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
)

type rafterTest struct {
	client dynamic.Interface
}

var (
	//_ BackupTest = NewRafterTest(nil)

	rafterWaitTimeout = 4 * time.Minute

	rafterAssetGroupName = "backup-asset-group"
	rafterAssetName      = "backup-asset"
	rafterBucketName     = "backup-bucket"

	rafterSourceURL = "https://raw.githubusercontent.com/kyma-project/kyma/master/README.md"
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "AssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
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
						Type: "sample",
						Name: "readme",
						Mode: v1beta1.AssetGroupSingle,
						URL:  rafterSourceURL,
					},
				},
			},
		},
	}

	return rt.create(v1beta1.GroupVersion.WithResource("assetgroups"), namespace, assetGroup)
}

func (rt *rafterTest) createBucket(namespace string) error {
	bucket := &v1beta1.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bucket",
			APIVersion: v1beta1.GroupVersion.String(),
		},
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

	return rt.create(v1beta1.GroupVersion.WithResource("buckets"), namespace, bucket)
}

func (rt *rafterTest) createAsset(namespace string) error {
	asset := &v1beta1.Asset{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Asset",
			APIVersion: v1beta1.GroupVersion.String(),
		},
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

	return rt.create(v1beta1.GroupVersion.WithResource("assets"), namespace, asset)
}

func (rt *rafterTest) testAssetGroup(namespace string) error {
	var assetGroup *v1beta1.AssetGroup
	err := waiter.WaitAtMost(func() (bool, error) {
		assetGroup = &v1beta1.AssetGroup{}
		err := rt.get(v1beta1.GroupVersion.WithResource("assetgroups"), namespace, rafterAssetGroupName, assetGroup)
		if err != nil {
			return false, err
		}

		if assetGroup.Status.Phase != v1beta1.AssetGroupReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		if assetGroup != nil {
			log.Printf("AssetGroup %s/%s test failed, phase: %s, reason [%s]: %s", namespace, rafterAssetGroupName, assetGroup.Status.Phase, assetGroup.Status.Reason, assetGroup.Status.Message)
		}
		return errors.Wrapf(err, "while waiting for ready AssetGroup resource")
	}

	return nil
}

func (rt *rafterTest) testBucket(namespace string) error {
	var bucket *v1beta1.Bucket
	err := waiter.WaitAtMost(func() (bool, error) {
		bucket = &v1beta1.Bucket{}
		err := rt.get(v1beta1.GroupVersion.WithResource("buckets"), namespace, rafterBucketName, bucket)
		if err != nil {
			return false, err
		}

		if bucket.Status.Phase != v1beta1.BucketReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		if bucket != nil {
			log.Printf("Bucket %s/%s test failed, phase %s,reason: [%s]: %s", namespace, rafterBucketName, bucket.Status.Phase, bucket.Status.Reason, bucket.Status.Message)
		}
		return errors.Wrapf(err, "while waiting for ready Bucket resource")
	}

	return nil
}

func (rt *rafterTest) testAsset(namespace string) error {
	var asset *v1beta1.Asset
	err := waiter.WaitAtMost(func() (bool, error) {
		asset = &v1beta1.Asset{}
		err := rt.get(v1beta1.GroupVersion.WithResource("assets"), namespace, rafterAssetName, asset)
		if err != nil {
			return false, err
		}

		if asset.Status.Phase != v1beta1.AssetReady {
			return false, nil
		}

		return true, nil
	}, rafterWaitTimeout)

	if err != nil {
		if asset != nil {
			log.Printf("Asset %s/%s test failed, phase %s, reason: [%s]: %s", namespace, rafterAssetGroupName, asset.Status.Phase, asset.Status.Reason, asset.Status.Message)
		}
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
