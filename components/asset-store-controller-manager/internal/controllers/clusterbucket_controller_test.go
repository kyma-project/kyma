package controllers

import (
	"context"
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

var _ = Describe("ClusterBucket", func() {
	var (
		bucket     *assetstorev1alpha2.ClusterBucket
		reconciler *ClusterBucketReconciler
		mocks      *MockContainer
		t          GinkgoTInterface
		request    ctrl.Request
	)

	BeforeEach(func() {
		bucket = newFixClusterBucket()
		Expect(k8sClient.Create(context.TODO(), bucket)).To(Succeed())
		t = GinkgoT()
		mocks = NewMockContainer()

		request = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      bucket.Name,
				Namespace: bucket.Namespace,
			},
		}

		reconciler = &ClusterBucketReconciler{
			Client:                  k8sClient,
			cacheSynchronizer:       func(stop <-chan struct{}) bool { return true },
			Log:                     log.Log,
			recorder:                record.NewFakeRecorder(100),
			relistInterval:          60 * time.Hour,
			store:                   mocks.Store,
			finalizer:               finalizer.New(deleteAssetFinalizerName),
			externalEndpoint:        "https://minio.test.local",
			maxConcurrentReconciles: 1,
		}
	})

	AfterEach(func() {
		mocks.AssertExpetactions(t)
	})

	It("should successfully create, update and delete ClusterBucket", func() {
		By("creating the ClusterBucket")
		// given
		mocks.Store.On("CreateBucket", bucket.Namespace, bucket.Name, string(bucket.Spec.Region)).Return("test", nil).Once()
		mocks.Store.On("SetBucketPolicy", "test", bucket.Spec.Policy).Return(nil).Once()

		// when
		result, err := reconciler.Reconcile(request)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())
		Expect(result.RequeueAfter).To(Equal(60 * time.Hour))

		// when
		err = k8sClient.Get(context.TODO(), request.NamespacedName, bucket)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bucket.Status.Phase).To(Equal(assetstorev1alpha2.BucketReady))
		Expect(bucket.Status.Reason).To(Equal(assetstorev1alpha2.BucketPolicyUpdated))

		By("updating the ClusterBucket")
		// when
		bucket.Spec.Policy = assetstorev1alpha2.BucketPolicyNone
		err = k8sClient.Update(context.TODO(), bucket)
		// then
		Expect(err).ToNot(HaveOccurred())

		// given
		mocks.Store.On("BucketExists", "test").Return(true, nil).Once()
		mocks.Store.On("CompareBucketPolicy", "test", bucket.Spec.Policy).Return(false, nil).Once()
		mocks.Store.On("SetBucketPolicy", "test", bucket.Spec.Policy).Return(nil).Once()

		// when
		result, err = reconciler.Reconcile(request)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())
		Expect(result.RequeueAfter).To(Equal(60 * time.Hour))

		// when
		err = k8sClient.Get(context.TODO(), request.NamespacedName, bucket)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(bucket.Status.Phase).To(Equal(assetstorev1alpha2.BucketReady))
		Expect(bucket.Status.Reason).To(Equal(assetstorev1alpha2.BucketPolicyUpdated))

		By("deleting the ClusterBucket")
		// when
		err = k8sClient.Delete(context.TODO(), bucket)
		// then
		Expect(err).ToNot(HaveOccurred())

		// given
		mocks.Store.On("DeleteBucket", mock.Anything, "test").Return(nil).Once()

		// when
		result, err = reconciler.Reconcile(request)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Requeue).To(BeFalse())
		Expect(result.RequeueAfter).To(Equal(60 * time.Hour))

		// when
		err = k8sClient.Get(context.TODO(), request.NamespacedName, bucket)
		// then
		Expect(err).To(HaveOccurred())
		Expect(apiErrors.IsNotFound(err)).To(BeTrue())
	})
})

func newFixClusterBucket() *assetstorev1alpha2.ClusterBucket {
	return &assetstorev1alpha2.ClusterBucket{
		ObjectMeta: ctrl.ObjectMeta{
			Name: string(uuid.NewUUID()),
		},
		Spec: assetstorev1alpha2.ClusterBucketSpec{
			CommonBucketSpec: assetstorev1alpha2.CommonBucketSpec{
				Region: assetstorev1alpha2.BucketRegionAPNortheast1,
				Policy: assetstorev1alpha2.BucketPolicyReadOnly,
			},
		},
		Status: assetstorev1alpha2.ClusterBucketStatus{CommonBucketStatus: assetstorev1alpha2.CommonBucketStatus{
			LastHeartbeatTime: v1.Now(),
		}},
	}
}
