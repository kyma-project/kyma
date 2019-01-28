package bucket

import (
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler/automock"
	"github.com/pkg/errors"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
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
		accessKeyName := "APP_ACCESS_KEY"
		secretKeyName := "APP_SECRET_KEY"
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
	regionName := string(assetstorev1alpha1.BucketRegionUSEast1)

	instance := fixInitialBucket(name, namespace, regionName, "")

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, regionName).Return(true, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "").Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
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

	bucket := &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Finalizers).To(gomega.ContainElement(DeleteBucketFinalizerName))
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketCreated"))
}

func TestReconcileBucketCreationFailed(t *testing.T) {
	// Given
	name := "bucket-creation-failed"
	exp := expectedFor(name, namespace)

	instance := fixInitialBucket(name, namespace, "", "")

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(false, testErr).Once()
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(true, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "").Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket := &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketCreationFailure"))
	// Now creating bucket and setting policy will pass
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
}

func TestReconcileBucketCheckFailed(t *testing.T) {
	// Given
	name := "bucket-check-failed"
	exp := expectedFor(name, namespace)

	instance := fixReadyBucket(name, namespace)

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("Exists", exp.BucketName).Return(false, testErr).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, instance.Spec.Policy).Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	// should retry checking bucket
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
}

func TestReconcileBucketPolicyUpdateSuccess(t *testing.T) {
	// Given
	name := "bucket-success-policy"
	exp := expectedFor(name, namespace)

	bucket := fixInitialBucket(name, namespace, "", "policy1")
	expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(true, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Times(4)
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy1").Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy1").Return(false, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy2").Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy2").Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
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
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	bucket = &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdated"))

	// When
	bucket = &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	bucket.Spec.Policy = "policy2"
	err = c.Update(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	bucket = &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdated"))
}

func TestReconcileBucketUpdatePolicyFailed(t *testing.T) {
	// Given
	name := "bucket-failed-policy"
	exp := expectedFor(name, namespace)

	instance := fixInitialBucket(name, namespace, "", "policy1")
	expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(true, nil).Once()
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(false, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy1").Return(false, testErr).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "policy1").Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, exp, instance)

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	bucket := &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdateFailed"))
}

func TestReconcileBucketDeletedRemotely(t *testing.T) {
	// Given
	name := "bucket-deleted-remotely"
	exp := expectedFor(name, namespace)

	regionName := string(assetstorev1alpha1.BucketRegionUSEast1)
	bucket := fixInitialBucket(name, namespace, regionName, "")

	bucketHandler := &automock.BucketHandler{}
	bucketHandlerBefore := bucketHandler
	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, regionName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "").Return(false, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("Exists", exp.BucketName).Return(false, nil).Twice()
	defer bucketHandlerBefore.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
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

	bucket = &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))

	//When
	bucket.Labels = map[string]string{"test": "label"}
	err = c.Update(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	bucket = &assetstorev1alpha1.Bucket{}
	err = c.Get(context.TODO(), exp.Key, bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketFailed))
	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketNotFound"))
}

func TestReconcileBucketDeleteFailed(t *testing.T) {
	// Given
	name := "bucket-delete-failed"
	exp := expectedFor(name, namespace)

	instance := fixReadyBucket(name, namespace)

	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, instance.Spec.Policy).Return(false, nil).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(testErr).Once()
	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	err = c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	//Then
	g.Eventually(func() bool {
		bucket := &assetstorev1alpha1.Bucket{}
		err := c.Get(context.TODO(), exp.Key, bucket)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}

func TestReconcileBucketAlreadyWithoutFinalizer(t *testing.T) {
	// Given
	name := "bucket-delete-failed"
	exp := expectedFor(name, namespace)

	instance := fixReadyBucket(name, namespace)
	instance.Finalizers = []string{}
	bucketHandler := &automock.BucketHandler{}
	bucketHandler.On("Exists", exp.BucketName).Return(true, nil).Once()
	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, instance.Spec.Policy).Return(false, nil).Once()
	defer bucketHandler.AssertExpectations(t)

	cfg := prepareReconcilerTest(t, bucketHandler)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	// When
	err := c.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	err = c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))

	//Then
	g.Eventually(func() bool {
		bucket := &assetstorev1alpha1.Bucket{}
		err := c.Get(context.TODO(), exp.Key, bucket)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}

func fixInitialBucket(name, namespace, region string, policy string) *assetstorev1alpha1.Bucket {
	return &assetstorev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: assetstorev1alpha1.BucketSpec{
			Region: assetstorev1alpha1.BucketRegion(region),
			Policy: policy,
		},
		Status: assetstorev1alpha1.BucketStatus{
			LastHeartbeatTime: metav1.Now(),
		},
	}
}

func fixReadyBucket(name, namespace string) *assetstorev1alpha1.Bucket {
	return &assetstorev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Finalizers: []string{
				DeleteBucketFinalizerName,
			},
		},
		Spec: assetstorev1alpha1.BucketSpec{
			Region: "",
			Policy: "",
		},
		Status: assetstorev1alpha1.BucketStatus{
			LastHeartbeatTime: metav1.Now(),
			Phase:             assetstorev1alpha1.BucketReady,
			Reason:            "BucketCreated",
			Message:           "Bucket has been successfully created",
		},
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

func deleteAndExpectSuccess(cfg *testSuite, exp expected, instance *assetstorev1alpha1.Bucket) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	g.Eventually(func() bool {
		instance := &assetstorev1alpha1.Bucket{}
		err := c.Get(context.TODO(), exp.Key, instance)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}
