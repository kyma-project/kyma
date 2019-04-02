package bucket

import (
	"fmt"
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
const namespace = "default"

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

func TestReconcileBucketCreationSuccess(t *testing.T) {
	// Given
	name := "bucket-creation-success"
	exp := expectedFor(name, namespace)
	regionName := string(assetstorev1alpha2.BucketRegionUSEast1)

	instance := fixInitialBucket(name, namespace, regionName, assetstorev1alpha2.BucketPolicyReadOnly)

	store := new(automock.Store)
	store.On("CreateBucket", namespace, name, regionName).Return(exp.BucketName, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, assetstorev1alpha2.BucketPolicyReadOnly).Return(nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket := &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Finalizers).To(gomega.ContainElement(deleteBucketFinalizerName))
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
}

func TestReconcileBucketCreationFailed(t *testing.T) {
	// Given
	name := "bucket-creation-failed"
	exp := expectedFor(name, namespace)

	instance := fixInitialBucket(name, namespace, "", assetstorev1alpha2.BucketPolicy(""))

	store := new(automock.Store)
	store.On("CreateBucket", namespace, name, "").Return(exp.BucketName, testErr).Once()
	store.On("CreateBucket", namespace, name, "").Return(exp.BucketName, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket := &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketCreationFailure"))

	// Now creating bucket and setting policy will pass
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdated"))

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
}

func TestReconcileBucketCheckFailed(t *testing.T) {
	// Given
	name := "bucket-check-failed-dupa"
	exp := expectedFor(name, namespace)

	instance := fixReadyBucket(name, namespace)

	store := new(automock.Store)
	store.On("BucketExists", exp.BucketName).Return(false, testErr).Once()
	store.On("BucketExists", exp.BucketName).Return(true, nil).Once()
	store.On("CompareBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(false, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)
	err = c.Status().Update(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	// should retry checking bucket
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
}

func TestReconcileBucketPolicyUpdateSuccess(t *testing.T) {
	// Given
	name := "bucket-success-policy"
	exp := expectedFor(name, namespace)

	bucket := fixInitialBucket(name, namespace, "", assetstorev1alpha2.BucketPolicyReadOnly)
	expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}

	store := new(automock.Store)
	store.On("CreateBucket", namespace, name, "").Return(exp.BucketName, nil).Once()
	store.On("BucketExists", exp.BucketName).Return(true, nil).Once()
	store.On("CompareBucketPolicy", exp.BucketName, assetstorev1alpha2.BucketPolicyWriteOnly).Return(false, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, bucket.Spec.Policy).Return(nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, assetstorev1alpha2.BucketPolicyWriteOnly).Return(nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, bucket)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))

	// When
	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	bucket.Spec.Policy = assetstorev1alpha2.BucketPolicyWriteOnly
	err = c.Update(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
}

func TestReconcileBucketUpdatePolicyFailed(t *testing.T) {
	// Given
	name := "bucket-failed-policy"
	exp := expectedFor(name, namespace)

	instance := fixInitialBucket(name, namespace, "", assetstorev1alpha2.BucketPolicyWriteOnly)
	expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}

	store := new(automock.Store)
	store.On("BucketExists", exp.BucketName).Return(true, nil).Once()
	store.On("CreateBucket", namespace, name, "").Return(exp.BucketName, nil).Once()
	store.On("CompareBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(false, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(testErr).Once()
	store.On("SetBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	bucket := &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdateFailed.String()))

	// Now setting policy will pass
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdated"))

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
}

func TestReconcileBucketDeletedRemotely(t *testing.T) {
	// Given
	name := "bucket-deleted-remotely"
	exp := expectedFor(name, namespace)

	regionName := string(assetstorev1alpha2.BucketRegionUSEast1)
	bucket := fixInitialBucket(name, namespace, regionName, assetstorev1alpha2.BucketPolicyReadOnly)

	store := new(automock.Store)
	store.On("BucketExists", exp.BucketName).Return(false, nil).Once()
	store.On("CreateBucket", namespace, name, regionName).Return(exp.BucketName, nil).Once()
	store.On("SetBucketPolicy", exp.BucketName, bucket.Spec.Policy).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, bucket)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketReady))

	//When
	bucket.Spec.Policy = assetstorev1alpha2.BucketPolicyWriteOnly
	err = c.Update(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha2.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha2.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketNotFound"))
}

func TestReconcileBucketDeleteFailed(t *testing.T) {
	// Given
	name := "bucket-delete-failed"
	exp := expectedFor(name, namespace)

	instance := fixReadyBucket(name, namespace)

	store := new(automock.Store)
	store.On("BucketExists", exp.BucketName).Return(true, nil).Once()
	store.On("CompareBucketPolicy", exp.BucketName, instance.Spec.Policy).Return(true, nil).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(testErr).Once()
	store.On("DeleteBucket", mock.Anything, exp.BucketName).Return(nil).Once()
	defer store.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, store)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	err = c.Status().Update(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	err = c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	//Then
	g.Eventually(func() bool {
		bucket := &assetstorev1alpha2.Bucket{}
		err := c.Get(context.TODO(), exp.Key, bucket)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}

func fixInitialBucket(name, namespace, region string, policy assetstorev1alpha2.BucketPolicy) *assetstorev1alpha2.Bucket {
	return &assetstorev1alpha2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: assetstorev1alpha2.BucketSpec{CommonBucketSpec: assetstorev1alpha2.CommonBucketSpec{
			Region: assetstorev1alpha2.BucketRegion(region),
			Policy: policy,
		}},
		Status: assetstorev1alpha2.BucketStatus{CommonBucketStatus: assetstorev1alpha2.CommonBucketStatus{
			LastHeartbeatTime: metav1.Now(),
		}},
	}
}

func fixReadyBucket(name, namespace string) *assetstorev1alpha2.Bucket {
	return &assetstorev1alpha2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Finalizers: []string{
				deleteBucketFinalizerName,
			},
		},
		Spec: assetstorev1alpha2.BucketSpec{CommonBucketSpec: assetstorev1alpha2.CommonBucketSpec{
			Region: "",
			Policy: "",
		}},
		Status: assetstorev1alpha2.BucketStatus{CommonBucketStatus: assetstorev1alpha2.CommonBucketStatus{
			RemoteName:         fmt.Sprintf("ns-%s-%s", namespace, name),
			ObservedGeneration: int64(1),
			LastHeartbeatTime:  metav1.NewTime(time.Now().Add(-61 * time.Hour)),
			Phase:              assetstorev1alpha2.BucketReady,
			Reason:             "BucketCreated",
			Message:            "Bucket has been successfully created",
		}},
	}
}

type expected struct {
	BucketName string
	Key        types.NamespacedName
	Request    reconcile.Request
}

func expectedFor(name, namespace string) expected {
	return expected{
		BucketName: fmt.Sprintf("ns-%s-%s", namespace, name),
		Key:        types.NamespacedName{Name: name, Namespace: namespace},
		Request:    reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}},
	}
}

func deleteWithoutReconcile(cfg *testSuite, instance *assetstorev1alpha2.Bucket) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func deleteAndExpectSuccess(cfg *testSuite, exp expected, instance *assetstorev1alpha2.Bucket) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(func() bool {
		instance := &assetstorev1alpha2.Bucket{}
		err := c.Get(context.TODO(), exp.Key, instance)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}
