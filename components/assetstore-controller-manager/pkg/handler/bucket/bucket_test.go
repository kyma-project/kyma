package bucket_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/bucket"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/bucket/pretty"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store/automock"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
	"time"
)

var log = logf.Log.WithName("asset-test")

func TestBucketHandler_IsOnAddOrUpdate(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnAddOrUpdate(testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Updated", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.ObjectMeta.Generation = int64(2)
		testData.Status.ObservedGeneration = int64(1)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnAddOrUpdate(testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("NotChanged", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.ObjectMeta.Generation = int64(1)
		testData.Status.ObservedGeneration = int64(1)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnAddOrUpdate(testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})
}

func TestBucketHandler_IsOnDelete(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		deletionTimestamp := v1.Now()
		testData.ObjectMeta.DeletionTimestamp = &deletionTimestamp
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnDelete(testData)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("False", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnDelete(testData)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})
}

func TestBucketHandler_IsOnFailed(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnFailed(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Ready", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.Phase = v1alpha2.BucketReady
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnFailed(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("True", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.Phase = v1alpha2.BucketFailed
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnFailed(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func TestBucketHandler_IsOnReady(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnReady(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Failed", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.Phase = v1alpha2.BucketFailed
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnReady(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("True", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.Phase = v1alpha2.BucketReady
		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.IsOnReady(testData.Status.CommonBucketStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func TestBucketHandler_OnAddOrUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", testData.Namespace, testData.Name, string(testData.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnAddOrUpdate(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("SuccessUpdateCreated", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(true, nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnAddOrUpdate(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("SuccessUpdateNotNotCreated", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", testData.Namespace, testData.Name, string(testData.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnAddOrUpdate(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})
}

func TestBucketHandler_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.RemoteName = "buckecior-123"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("DeleteBucket", ctx, testData.Status.RemoteName).Return(nil).Once()
		defer store.AssertExpectations(t)

		bucketHandler := bucket.New(fakeRecorder(), store, "", log)

		// When
		err := bucketHandler.OnDelete(ctx, testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("NoRemoteName", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		bucketHandler := bucket.New(fakeRecorder(), store, "", log)

		// When
		err := bucketHandler.OnDelete(ctx, testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.RemoteName = "buckecior-123"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("DeleteBucket", ctx, testData.Status.RemoteName).Return(errors.New("err")).Once()
		defer store.AssertExpectations(t)

		bucketHandler := bucket.New(fakeRecorder(), store, "", log)

		// When
		err := bucketHandler.OnDelete(ctx, testData, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestBucketHandler_OnFailed(t *testing.T) {
	t.Run("OnBucketCreationFailure", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.Phase = v1alpha2.BucketFailed
		testData.Status.Reason = pretty.BucketCreationFailure.String()
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", testData.Namespace, testData.Name, string(testData.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status, err := bucketHandler.OnFailed(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("BucketVerificationFailure", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)
		testData.Status.Phase = v1alpha2.BucketFailed
		testData.Status.Reason = pretty.BucketVerificationFailure.String()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(true, nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status, err := bucketHandler.OnFailed(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("BucketPolicyUpdateFailed", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)
		testData.Status.Phase = v1alpha2.BucketFailed
		testData.Status.Reason = pretty.BucketPolicyUpdateFailed.String()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status, err := bucketHandler.OnFailed(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("StillFailingWithSameReason", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)
		testData.Status.Phase = v1alpha2.BucketFailed
		testData.Status.Reason = pretty.BucketVerificationFailure.String()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(false, errors.New("test-err")).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		_, err := bucketHandler.OnFailed(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestBucketHandler_OnReady(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(true, nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnReady(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("MissingBucket", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(false, nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnReady(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.NotFoundReason.String()))
	})

	t.Run("ErrorOnBucketVerification", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(false, errors.New("test-message")).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnReady(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketVerificationFailure.String()))
	})

	t.Run("InvalidPolicy", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(nil).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnReady(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdated.String()))
		g.Expect(status.RemoteName).To(gomega.Equal(remoteName))
		g.Expect(status.Url).To(gomega.Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("PolicyError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		url := "http://trtrt.com"
		remoteName := fmt.Sprintf("%s-123", testData.Name)
		testData.Status.RemoteName = remoteName
		testData.Status.Url = fmt.Sprintf("%s/%s", url, remoteName)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", remoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", remoteName, testData.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", remoteName, testData.Spec.Policy).Return(errors.New("test-err")).Once()

		bucketHandler := bucket.New(fakeRecorder(), store, url, log)

		// When
		status := bucketHandler.OnReady(testData, testData.Spec.CommonBucketSpec, testData.Status.CommonBucketStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketPolicyUpdateFailed.String()))
	})
}

func TestBucketHandler_ShouldReconcile(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		relistInterval := time.Second

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Updated", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		testData.ObjectMeta.Generation = int64(2)
		testData.Status.ObservedGeneration = int64(1)
		relistInterval := time.Second

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("BeingDeleted", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		testData.ObjectMeta.Generation = int64(1)
		testData.Status.ObservedGeneration = int64(1)
		testData.ObjectMeta.DeletionTimestamp = &testData.Status.LastHeartbeatTime
		relistInterval := time.Second

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Ready", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		relistInterval := time.Second
		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now.Add(-10 * relistInterval))
		testData.Status.Phase = v1alpha2.BucketReady
		testData.ObjectMeta.Generation = int64(1)
		testData.Status.ObservedGeneration = int64(1)

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("ReadySkipped", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		relistInterval := time.Second
		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now.Add(relistInterval))
		testData.Status.Phase = v1alpha2.BucketReady
		testData.ObjectMeta.Generation = int64(1)
		testData.Status.ObservedGeneration = int64(1)

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Default", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		relistInterval := time.Second
		now := time.Now()
		testData := testData("new", v1alpha2.BucketPolicyReadOnly)
		testData.Status.LastHeartbeatTime = v1.NewTime(now.Add(relistInterval))
		testData.ObjectMeta.Generation = int64(1)
		testData.Status.ObservedGeneration = int64(1)
		testData.Status.Phase = v1alpha2.BucketFailed

		bucketHandler := bucket.New(nil, nil, "", log)

		// When
		result := bucketHandler.ShouldReconcile(testData, testData.Status.CommonBucketStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func fakeRecorder() record.EventRecorder {
	return record.NewFakeRecorder(20)
}

func testData(name string, policy v1alpha2.BucketPolicy) *v1alpha2.Bucket {
	return &v1alpha2.Bucket{
		ObjectMeta: v1.ObjectMeta{
			Name:       name,
			Namespace:  fmt.Sprintf("%s-ns", name),
			Generation: int64(1),
		},
		Spec: v1alpha2.BucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy: policy,
				Region: "asia",
			},
		},
	}
}
