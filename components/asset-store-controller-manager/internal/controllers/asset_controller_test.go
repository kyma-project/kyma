package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/finalizer"
	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var _ = Describe("Asset", func() {
	var (
		asset      *assetstorev1alpha2.Asset
		baseURL    string
		reconciler *AssetReconciler
		mocks      *MockContainer
		t          GinkgoTInterface
		request    ctrl.Request
	)

	BeforeEach(func() {
		bucket := newFixBucket()
		bucket.Status.Phase = assetstorev1alpha2.BucketReady
		bucket.Status.RemoteName = bucket.Name
		Expect(k8sClient.Create(context.TODO(), bucket)).To(Succeed())
		Expect(k8sClient.Status().Update(context.TODO(), bucket)).To(Succeed())

		asset = newFixAsset(bucket.Name)
		Expect(k8sClient.Create(context.TODO(), asset)).To(Succeed())
		baseURL = fmt.Sprintf("%s/%s", bucket.Status.URL, asset.Name)

		t = GinkgoT()
		mocks = NewMockContainer()

		request = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      asset.Name,
				Namespace: asset.Namespace,
			},
		}

		reconciler = &AssetReconciler{
			Client:                  k8sClient,
			cacheSynchronizer:       func(stop <-chan struct{}) bool { return true },
			Log:                     log.Log,
			recorder:                record.NewFakeRecorder(100),
			relistInterval:          60 * time.Hour,
			store:                   mocks.Store,
			loader:                  mocks.Loader,
			finalizer:               finalizer.New(deleteAssetFinalizerName),
			validator:               mocks.Validator,
			mutator:                 mocks.Mutator,
			metadataExtractor:       mocks.Extractor,
			maxConcurrentReconciles: 1,
		}
	})

	AfterEach(func() {
		mocks.AssertExpetactions(t)
	})

	It("should successfully create, update and delete Asset", func() {
		By("creating the Asset")
		// On scheduled
		result, err := reconciler.Reconcile(request)
		validateReconcilation(err, result)
		asset = &assetstorev1alpha2.Asset{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, asset)).To(Succeed())
		validateAsset(asset.Status.CommonAssetStatus, asset.ObjectMeta, "", []string{}, assetstorev1alpha2.AssetPending, assetstorev1alpha2.AssetScheduled)

		// On pending
		mocks.Store.On("ListObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name).Return([]string{}, nil).Once()
		mocks.Loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", []string{"test.file1", "test.file2"}, nil).Once()
		mocks.Loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.Store.On("PutObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name, "/tmp", []string{"test.file1", "test.file2"}).Return(nil).Once()

		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)
		asset = &assetstorev1alpha2.Asset{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, asset)).To(Succeed())
		validateAsset(asset.Status.CommonAssetStatus, asset.ObjectMeta, baseURL, []string{"test.file1", "test.file2"}, assetstorev1alpha2.AssetReady, assetstorev1alpha2.AssetUploaded)

		By("updating the Asset")
		asset.Spec.Source.URL = "example.com/test.file"
		asset.Spec.Source.Mode = assetstorev1alpha2.AssetSingle
		Expect(k8sClient.Update(context.TODO(), asset)).To(Succeed())

		// On scheduled
		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)
		asset = &assetstorev1alpha2.Asset{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, asset)).To(Succeed())
		validateAsset(asset.Status.CommonAssetStatus, asset.ObjectMeta, "", []string{}, assetstorev1alpha2.AssetPending, assetstorev1alpha2.AssetScheduled)

		// On pending
		mocks.Store.On("ListObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name).Return([]string{"test.file1", "test.file2"}, nil).Once()
		mocks.Store.On("DeleteObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name).Return(nil).Once()
		mocks.Loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", []string{"test.file"}, nil).Once()
		mocks.Loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.Store.On("PutObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name, "/tmp", []string{"test.file"}).Return(nil).Once()

		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)
		asset = &assetstorev1alpha2.Asset{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, asset)).To(Succeed())
		validateAsset(asset.Status.CommonAssetStatus, asset.ObjectMeta, baseURL, []string{"test.file"}, assetstorev1alpha2.AssetReady, assetstorev1alpha2.AssetUploaded)

		By("deleting the Asset")
		Expect(k8sClient.Delete(context.TODO(), asset)).To(Succeed())

		// On delete
		mocks.Store.On("ListObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name).Return([]string{"test.file"}, nil).Once()
		mocks.Store.On("DeleteObjects", mock.Anything, asset.Spec.BucketRef.Name, asset.Name).Return(nil).Once()

		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)

		asset = &assetstorev1alpha2.Asset{}
		err = k8sClient.Get(context.TODO(), request.NamespacedName, asset)
		Expect(err).To(HaveOccurred())
		Expect(apiErrors.IsNotFound(err)).To(BeTrue())
	})
})

func newFixAsset(bucketName string) *assetstorev1alpha2.Asset {
	return &assetstorev1alpha2.Asset{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      string(uuid.NewUUID()),
			Namespace: "default",
		},
		Spec: assetstorev1alpha2.AssetSpec{
			CommonAssetSpec: assetstorev1alpha2.CommonAssetSpec{
				Source: assetstorev1alpha2.AssetSource{
					URL:  "example.com/test.zip",
					Mode: assetstorev1alpha2.AssetPackage,
				},
				BucketRef: assetstorev1alpha2.AssetBucketRef{
					Name: bucketName,
				},
			},
		},
		Status: assetstorev1alpha2.AssetStatus{CommonAssetStatus: assetstorev1alpha2.CommonAssetStatus{
			LastHeartbeatTime: v1.Now(),
		}},
	}
}
