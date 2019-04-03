package clusterbucket

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/bucket/pretty"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const timeout = time.Second * 10

var testErr = errors.New("Test")

func TestAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//given
		g := gomega.NewGomegaWithT(t)

		// Set test envs
		accessKeyName := "APP_STORE_ACCESS_KEY"
		secretKeyName := "APP_STORE_SECRET_KEY"
		originalAccessKey := os.Getenv(accessKeyName)
		originalSecretKey := os.Getenv(secretKeyName)

		err := os.Setenv(accessKeyName, "test")
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		err = os.Setenv(secretKeyName, "test")

		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		mgr, err := manager.New(cfg, manager.Options{})
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		err = Add(mgr)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		// Restore envs
		err = os.Setenv(accessKeyName, originalAccessKey)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
		err = os.Setenv(secretKeyName, originalSecretKey)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		err := Add(nil)
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestReconcileClusterBucket_Reconcile(t *testing.T) {
	// Given
	name := "bucket-creation-success"
	exp := expectedFor(name)
	regionName := string(assetstorev1alpha2.BucketRegionUSEast1)

	instance := fixInitialBucket(name, regionName, assetstorev1alpha2.BucketPolicyReadOnly)
	updatedPolicy := assetstorev1alpha2.BucketPolicyReadWrite

	store := new(automock.Store)
	store.On("CreateBucket", "", name, regionName).Return(exp.BucketName, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, assetstorev1alpha2.BucketPolicyReadOnly).Return(nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, updatedPolicy).Return(nil).Once()
	store.On("BucketExists", name).Return(true, nil).Once()
	store.On("CompareBucketPolicy", name, updatedPolicy).Return(false, nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When - Create
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket := &assetstorev1alpha2.ClusterBucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	validateBucket(cfg, bucket, assetstorev1alpha2.BucketReady, pretty.BucketPolicyUpdated)

	// When - Update
	updated := bucket.DeepCopy()
	updated.Spec.Policy = updatedPolicy
	err = c.Update(context.TODO(), updated)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.ClusterBucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	validateBucket(cfg, bucket, assetstorev1alpha2.BucketReady, pretty.BucketPolicyUpdated)

	// When - Delete
	err = c.Delete(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.ClusterBucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(apierrors.IsNotFound(err)).To(gomega.BeTrue())
}

func validateBucket(cfg *testSuite, instance *assetstorev1alpha2.ClusterBucket, phase assetstorev1alpha2.BucketPhase, reason pretty.Reason) {
	g := cfg.g
	g.Expect(instance.Finalizers).To(gomega.ContainElement(deleteBucketFinalizerName))
	g.Expect(instance.Status.Phase).To(gomega.Equal(phase))
	g.Expect(instance.Status.Reason).To(gomega.Equal(reason.String()))
}

func fixInitialBucket(name, region string, policy assetstorev1alpha2.BucketPolicy) *assetstorev1alpha2.ClusterBucket {
	return &assetstorev1alpha2.ClusterBucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: assetstorev1alpha2.ClusterBucketSpec{CommonBucketSpec: assetstorev1alpha2.CommonBucketSpec{
			Region: assetstorev1alpha2.BucketRegion(region),
			Policy: policy,
		}},
		Status: assetstorev1alpha2.ClusterBucketStatus{CommonBucketStatus: assetstorev1alpha2.CommonBucketStatus{
			LastHeartbeatTime: metav1.Now(),
		}},
	}
}

type expected struct {
	BucketName string
	Key        types.NamespacedName
	Request    reconcile.Request
}

func expectedFor(name string) expected {
	return expected{
		BucketName: name,
		Key:        types.NamespacedName{Name: name, Namespace: ""},
		Request:    reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: ""}},
	}
}
