package bucket

import (
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler/automock"
	"github.com/pkg/errors"
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

const timeout = time.Second * 30
//
//func TestAdd(t *testing.T) {
//	t.Run("LoadConfigError", func(t *testing.T) {
//		g := gomega.NewGomegaWithT(t)
//		mgr, err := manager.New(cfg, manager.Options{})
//		g.Expect(err).ShouldNot(gomega.HaveOccurred())
//		err = Add(mgr)
//		g.Expect(err).Should(gomega.HaveOccurred())
//	})
//}

func TestReconcile(t *testing.T) {
	namespace := "default"
	testErr := errors.New("Test")

	t.Run("BucketCreationSuccess", func(t *testing.T) {
		// Given
		name := "bucket-creation-success"
		exp := expectedFor(name, namespace)

		instance := fixInitialBucket(name, namespace, "test", "")

		bucketHandler := &automock.BucketHandler{}
		bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "test").Return(true, nil).Once()
		bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil).Once()
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
		g.Expect(bucket.Finalizers).To(gomega.ContainElement(DeleteBucketFinalizerName))
		g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
		g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketCreated"))
	})

	t.Run("BucketCreationFailed", func(t *testing.T) {
		// Given
		name := "bucket-creation-failed"
		exp := expectedFor(name, namespace)

		instance := fixInitialBucket(name, namespace, "", "")

		bucketHandler := &automock.BucketHandler{}
		bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(false, testErr)
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
	})

	t.Run("BucketCheckFailed", func(t *testing.T) {
		// Given
		name := "bucket-check-failed"
		exp := expectedFor(name, namespace)

		instance := fixReadyBucket(name, namespace)

		bucketHandler := &automock.BucketHandler{}
		bucketHandler.On("CheckIfExists", exp.BucketName).Return(false, testErr).Once()
		bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil)
		bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, string(instance.Spec.Policy)).Return(false, nil)
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
		// check updated heartbeat time
		bucket := &assetstorev1alpha1.Bucket{}
		err = c.Get(context.TODO(), exp.Key, bucket)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(bucket.Status.LastHeartbeatTime).NotTo(gomega.Equal(instance.Status.LastHeartbeatTime))
	})

	t.Run("BucketPolicyUpdateSuccess", func(t *testing.T) {
		// Given
		name := "bucket-success-policy"
		exp := expectedFor(name, namespace)

		bucket := fixInitialBucket(name, namespace, "", "readonly")
		expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}

		bucketHandler := &automock.BucketHandler{}
		bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(true, nil).Once()
		bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil)
		bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "readonly").Return(true, nil)
		bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "readwrite").Return(true, nil).Once()
		bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "readwrite").Return(false, nil)
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

		bucket = &assetstorev1alpha1.Bucket{}
		err = c.Get(context.TODO(), exp.Key, bucket)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
		g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdated"))

		// When
		bucket = &assetstorev1alpha1.Bucket{}
		err = c.Get(context.TODO(), exp.Key, bucket)
		g.Expect(err).NotTo(gomega.HaveOccurred())
		bucket.Spec.Policy = "readwrite"
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
	})

	//t.Run("BucketUpdatePolicyFailed", func(t *testing.T) {
	//	// Given
	//	name := "bucket-failed-policy"
	//	exp := expectedFor(name, namespace)
	//
	//	instance := fixInitialBucket(name, namespace, "", "readonly")
	//	expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}
	//
	//	bucketHandler := &automock.BucketHandler{}
	//	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "").Return(true, nil).Once()
	//	bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil)
	//	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "readonly").Return(false, testErr)
	//	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	//	defer bucketHandler.AssertExpectations(t)
	//
	//	cfg := prepareReconcilerTest(t, bucketHandler)
	//	g := cfg.g
	//	c := cfg.c
	//	defer cfg.finishTest()
	//
	//	// When
	//	err := c.Create(context.TODO(), instance)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//	defer deleteAndExpectSuccess(cfg, exp, instance)
	//
	//	// Then
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//
	//	bucket := &assetstorev1alpha1.Bucket{}
	//	err = c.Get(context.TODO(), exp.Key, bucket)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketFailed))
	//	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketPolicyUpdateFailed"))
	//})

	//t.Run("BucketDeletedRemotely", func(t *testing.T) {
	//	// Given
	//	name := "bucket-deleted-remotely"
	//	exp := expectedFor(name, namespace)
	//
	//	bucket := fixInitialBucket(name, namespace, "test", "")
	//
	//	bucketHandler := &automock.BucketHandler{}
	//	bucketHandlerBefore := bucketHandler
	//	bucketHandler.On("CreateIfDoesntExist", exp.BucketName, "test").Return(true, nil).Once()
	//	bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil).Once()
	//	bucketHandler.On("CheckIfExists", exp.BucketName).Return(false, nil)
	//	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, "").Return(false, nil)
	//	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	//	defer bucketHandlerBefore.AssertExpectations(t)
	//
	//	cfg := prepareReconcilerTest(t, bucketHandler)
	//	g := cfg.g
	//	c := cfg.c
	//	defer cfg.finishTest()
	//
	//	// When
	//	err := c.Create(context.TODO(), bucket)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//	defer deleteAndExpectSuccess(cfg, exp, bucket)
	//
	//	// Then
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	//
	//	bucket = &assetstorev1alpha1.Bucket{}
	//	err = c.Get(context.TODO(), exp.Key, bucket)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketReady))
	//
	//	//When
	//	bucket.Labels = map[string]string{"test": "label"}
	//	err = c.Update(context.TODO(), bucket)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//
	//	// Then
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	//	bucket = &assetstorev1alpha1.Bucket{}
	//	err = c.Get(context.TODO(), exp.Key, bucket)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//	g.Expect(bucket.Status.Phase).To(gomega.Equal(assetstorev1alpha1.BucketFailed))
	//	g.Expect(bucket.Status.Reason).To(gomega.Equal("BucketNotFound"))
	//})
	//
	//t.Run("BucketDeleteFailed", func(t *testing.T) {
	//	// Given
	//	name := "bucket-delete-failed"
	//	exp := expectedFor(name, namespace)
	//
	//	instance := fixReadyBucket(name, namespace)
	//
	//	bucketHandler := &automock.BucketHandler{}
	//	bucketHandler.On("CheckIfExists", exp.BucketName).Return(true, nil)
	//	bucketHandler.On("SetPolicyIfNotEqual", exp.BucketName, string(instance.Spec.Policy)).Return(false, nil)
	//	bucketHandler.On("Delete", exp.BucketName).Return(testErr).Once()
	//	bucketHandler.On("Delete", exp.BucketName).Return(nil).Once()
	//	defer bucketHandler.AssertExpectations(t)
	//
	//	cfg := prepareReconcilerTest(t, bucketHandler)
	//	g := cfg.g
	//	c := cfg.c
	//	defer cfg.finishTest()
	//
	//	// When
	//	err := c.Create(context.TODO(), instance)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//
	//	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(exp.Request)))
	//
	//	err = c.Delete(context.TODO(), instance)
	//	g.Expect(err).NotTo(gomega.HaveOccurred())
	//
	//	g.Eventually(func() bool {
	//		bucket := &assetstorev1alpha1.Bucket{}
	//		err := c.Get(context.TODO(), exp.Key, bucket)
	//		return apierrors.IsNotFound(err)
	//	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
	//})
}

func fixInitialBucket(name, namespace, region string, policy assetstorev1alpha1.BucketPolicy) *assetstorev1alpha1.Bucket {
	return &assetstorev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: assetstorev1alpha1.BucketSpec{
			Region: region,
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
