package bucket_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/handler/bucket"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/store/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("asset-test")

func TestBucketHandler_Handle_Default(t *testing.T) {
	// Given
	g := NewGomegaWithT(t)
	ctx := context.TODO()
	relistInterval := time.Minute
	now := time.Now()
	data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
	data.ObjectMeta.Generation = int64(1)
	data.Status.ObservedGeneration = int64(1)

	store := new(automock.Store)
	defer store.AssertExpectations(t)

	handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

	// When
	status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

	// Then
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(status).To(BeZero())
}

func TestBucketHandler_Handle_OnAddOrUpdate(t *testing.T) {
	t.Run("PolicyModified", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(2)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
	})

	t.Run("NoBucket", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(2)
		remoteName := fmt.Sprintf("%s-123", data.Name)
		url := "http://localhost"

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", data.Namespace, data.Name, string(data.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, data.Spec.Policy).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, url, relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
		g.Expect(status.RemoteName).To(Equal(remoteName))
		g.Expect(status.URL).To(Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("BucketCreationFailure", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(2)
		url := "http://localhost"

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", data.Namespace, data.Name, string(data.Spec.Region)).Return("", errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, url, relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketCreationFailure))
	})

	t.Run("BucketPolicyUpdateFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(2)
		remoteName := fmt.Sprintf("%s-123", data.Name)
		url := "http://localhost"

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", data.Namespace, data.Name, string(data.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, data.Spec.Policy).Return(errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, url, relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdateFailed))
		g.Expect(status.RemoteName).To(Equal(remoteName))
		g.Expect(status.URL).To(Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})
}

func TestBucketHandler_Handle_OnReady(t *testing.T) {
	t.Run("NotTaken", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.Now()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("NotChanged", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(true, nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
	})

	t.Run("MissingBucket", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(false, nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketNotFound))
	})

	t.Run("BucketVerificationFailure", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(false, errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketVerificationFailure))
	})

	t.Run("PolicyModified", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
	})

	t.Run("PolicyModificationError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(false, nil).Once()
		store.On("SetBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdateFailed))
	})

	t.Run("PolicyCompareError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketReady
		data.Status.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(false, errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketFailed))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyVerificationFailed))
	})
}

func TestBucketHandler_Handle_OnFailed(t *testing.T) {
	t.Run("BucketCreationFailure", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketFailed
		data.Status.Reason = v1alpha2.BucketCreationFailure
		remoteName := fmt.Sprintf("%s-123", data.Name)
		url := "http://localhost"

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", data.Namespace, data.Name, string(data.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, data.Spec.Policy).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, url, relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
		g.Expect(status.RemoteName).To(Equal(remoteName))
		g.Expect(status.URL).To(Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})

	t.Run("BucketVerificationFailure", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketFailed
		data.Status.Reason = v1alpha2.BucketVerificationFailure
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(true, nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
	})

	t.Run("BucketPolicyUpdateFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketFailed
		data.Status.Reason = v1alpha2.BucketPolicyUpdateFailed
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("BucketExists", data.Status.RemoteName).Return(true, nil).Once()
		store.On("CompareBucketPolicy", data.Status.RemoteName, data.Spec.Policy).Return(true, nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		data.ObjectMeta.Generation = int64(1)
		data.Status.ObservedGeneration = int64(1)
		data.Status.Phase = v1alpha2.BucketFailed
		data.Status.Reason = v1alpha2.BucketNotFound
		remoteName := fmt.Sprintf("%s-123", data.Name)
		url := "http://localhost"

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("CreateBucket", data.Namespace, data.Name, string(data.Spec.Region)).Return(remoteName, nil).Once()
		store.On("SetBucketPolicy", remoteName, data.Spec.Policy).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, url, relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.BucketReady))
		g.Expect(status.Reason).To(Equal(v1alpha2.BucketPolicyUpdated))
		g.Expect(status.RemoteName).To(Equal(remoteName))
		g.Expect(status.URL).To(Equal(fmt.Sprintf("%s/%s", url, remoteName)))
	})
}

func TestBucketHandler_Handle_OnDelete(t *testing.T) {
	t.Run("WithRemoteName", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		deletionTimestamp := v1.Now()
		data.ObjectMeta.DeletionTimestamp = &deletionTimestamp
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("DeleteBucket", ctx, data.Status.RemoteName).Return(nil).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("NoRemoteName", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		deletionTimestamp := v1.Now()
		data.ObjectMeta.DeletionTimestamp = &deletionTimestamp

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("BucketNotFound", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		deletionTimestamp := v1.Now()
		data.ObjectMeta.DeletionTimestamp = &deletionTimestamp
		data.Status.Reason = v1alpha2.BucketNotFound
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		data := testData("test-bucket", v1alpha2.BucketPolicyReadOnly)
		deletionTimestamp := v1.Now()
		data.ObjectMeta.DeletionTimestamp = &deletionTimestamp
		data.Status.RemoteName = fmt.Sprintf("%s-123", data.Name)

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		store.On("DeleteBucket", ctx, data.Status.RemoteName).Return(errors.New("nope")).Once()

		handler := bucket.New(log, fakeRecorder(), store, "https://localhost", relistInterval)

		// When
		status, err := handler.Do(ctx, now, data, data.Spec.CommonBucketSpec, data.Status.CommonBucketStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(BeZero())
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
