package asset_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	engineMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset/pretty"
	loaderMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/loader/automock"
	storeMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store/automock"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"testing"
	"time"
)

var log = logf.Log.WithName("asset-test")

const (
	remoteBucketName = "bucket-name"
)

func TestAssetHandler_Handle_OnAddOrUpdate(t *testing.T) {
	t.Run("OnAdd", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetPending))
		g.Expect(status.Reason).To(Equal(pretty.Scheduled.String()))
	})

	t.Run("OnUpdate", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.ObjectMeta.Generation = int64(2)
		asset.Status.ObservedGeneration = int64(1)

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetPending))
		g.Expect(status.Reason).To(Equal(pretty.Scheduled.String()))
	})
}

func TestAssetHandler_Handle_Default(t *testing.T) {
	// Given
	g := NewGomegaWithT(t)
	ctx := context.TODO()
	relistInterval := time.Minute
	now := time.Now()
	asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
	asset.ObjectMeta.Generation = int64(1)
	asset.Status.ObservedGeneration = int64(1)

	handler, mocks := newHandler(relistInterval)
	defer mocks.AssertExpectations(t)

	// When
	status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

	// Then
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(status).To(BeZero())
}

func TestAssetHandler_Handle_OnReady(t *testing.T) {
	t.Run("NotTaken", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now)
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

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
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ContainsAllObjects", ctx, remoteBucketName, asset.Name, mock.AnythingOfType("[]string")).Return(true, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetReady))
		g.Expect(status.Reason).To(Equal(pretty.Uploaded.String()))
	})

	t.Run("BucketNotReady", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "notReady", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetPending))
		g.Expect(status.Reason).To(Equal(pretty.BucketNotReady.String()))
	})

	t.Run("BucketError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "error", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.BucketError.String()))
	})

	t.Run("MissingFiles", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ContainsAllObjects", ctx, remoteBucketName, asset.Name, mock.AnythingOfType("[]string")).Return(false, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.MissingContent.String()))
	})

	t.Run("ContainsError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.NewTime(now.Add(-2 * relistInterval))
		asset.Status.CommonAssetStatus.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ContainsAllObjects", ctx, remoteBucketName, asset.Name, mock.AnythingOfType("[]string")).Return(false, errors.New("nope")).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.RemoteContentVerificationError.String()))
	})
}

func TestAssetHandler_Handle_OnPending(t *testing.T) {
	t.Run("WithWebhooks", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.store.On("PutObjects", ctx, remoteBucketName, asset.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(nil).Once()
		mocks.validator.On("Validate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetReady))
		g.Expect(status.Reason).To(Equal(pretty.Uploaded.String()))
	})

	t.Run("WithoutWebhooks", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation
		asset.Spec.Source.ValidationWebhookService = nil
		asset.Spec.Source.MutationWebhookService = nil

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.store.On("PutObjects", ctx, remoteBucketName, asset.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetReady))
		g.Expect(status.Reason).To(Equal(pretty.Uploaded.String()))
	})

	t.Run("LoadError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, errors.New("nope")).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.PullingFailed.String()))
	})

	t.Run("MutationFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(errors.New("nope")).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.MutationFailed.String()))
	})

	t.Run("ValidationError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(nil).Once()
		mocks.validator.On("Validate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: false}, errors.New("nope")).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.ValidationError.String()))
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(nil).Once()
		mocks.validator.On("Validate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: false}, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.ValidationFailed.String()))
	})

	t.Run("UploadError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.store.On("PutObjects", ctx, remoteBucketName, asset.Name, "/tmp", mock.AnythingOfType("[]string")).Return(errors.New("nope")).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(nil).Once()
		mocks.validator.On("Validate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.UploadFailed.String()))
	})

	t.Run("BucketNotReady", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "notReady", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetPending))
		g.Expect(status.Reason).To(Equal(pretty.BucketNotReady.String()))
	})

	t.Run("BucketStatusError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "error", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(Equal(pretty.BucketError.String()))
	})

	t.Run("OnBucketNotReadyBeforeTime", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "error", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		asset.Status.CommonAssetStatus.Reason = pretty.BucketNotReady.String()
		asset.Status.CommonAssetStatus.LastHeartbeatTime = v1.Now()
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})
}

func TestAssetHandler_Handle_OnFailed(t *testing.T) {
	t.Run("ShouldHandle", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetFailed
		asset.Status.CommonAssetStatus.Reason = pretty.BucketError.String()
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()
		mocks.store.On("PutObjects", ctx, remoteBucketName, asset.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		mocks.loader.On("Load", asset.Spec.Source.URL, asset.Name, asset.Spec.Source.Mode, asset.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		mocks.loader.On("Clean", "/tmp").Return(nil).Once()
		mocks.mutator.On("Mutate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.MutationWebhookService).Return(nil).Once()
		mocks.validator.On("Validate", ctx, asset, "/tmp", mock.AnythingOfType("[]string"), asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).ToNot(BeZero())
		g.Expect(status.Phase).To(Equal(v1alpha2.AssetReady))
		g.Expect(status.Reason).To(Equal(pretty.Uploaded.String()))
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetFailed
		asset.Status.CommonAssetStatus.Reason = pretty.ValidationFailed.String()
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("MutationFailed", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		asset.Status.CommonAssetStatus.Phase = v1alpha2.AssetFailed
		asset.Status.CommonAssetStatus.Reason = pretty.MutationFailed.String()
		asset.Status.ObservedGeneration = asset.Generation

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})
}

func TestAssetHandler_Handle_OnDelete(t *testing.T) {
	t.Run("NoFiles", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("MultipleFiles", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta
		files := []string{"test/a.txt", "test/b.txt"}

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(files, nil).Once()
		mocks.store.On("DeleteObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("ListError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(nil, errors.New("nope")).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "test-bucket", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta
		files := []string{"test/a.txt", "test/b.txt"}

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)
		mocks.store.On("ListObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(files, nil).Once()
		mocks.store.On("DeleteObjects", ctx, remoteBucketName, fmt.Sprintf("/%s", asset.Name)).Return(errors.New("nope")).Once()

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("BucketNotReady", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "notReady", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(status).To(BeZero())
	})

	t.Run("BucketStatusError", func(t *testing.T) {
		// Given
		g := NewGomegaWithT(t)
		ctx := context.TODO()
		relistInterval := time.Minute
		now := time.Now()
		asset := testData("test-asset", "error", "https://localhost/test.md")
		nowMeta := v1.Now()
		asset.ObjectMeta.DeletionTimestamp = &nowMeta

		handler, mocks := newHandler(relistInterval)
		defer mocks.AssertExpectations(t)

		// When
		status, err := handler.Do(ctx, now, asset, asset.Spec.CommonAssetSpec, asset.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(HaveOccurred())
		g.Expect(status).To(BeZero())
	})
}

func bucketStatusFinder(ctx context.Context, namespace, name string) (*v1alpha2.CommonBucketStatus, bool, error) {
	switch {
	case strings.Contains(name, "notReady"):
		return nil, false, nil
	case strings.Contains(name, "error"):
		return nil, false, errors.New("test-error")
	default:
		return &v1alpha2.CommonBucketStatus{
			Phase:      v1alpha2.BucketReady,
			URL:        "http://test-url.com/bucket-name",
			RemoteName: remoteBucketName,
		}, true, nil
	}
}

type mocks struct {
	store     *storeMock.Store
	loader    *loaderMock.Loader
	validator *engineMock.Validator
	mutator   *engineMock.Mutator
}

func (m *mocks) AssertExpectations(t *testing.T) {
	m.store.AssertExpectations(t)
	m.loader.AssertExpectations(t)
	m.validator.AssertExpectations(t)
	m.mutator.AssertExpectations(t)
}

func newHandler(relistInterval time.Duration) (asset.Handler, mocks) {
	mocks := mocks{
		store:     new(storeMock.Store),
		loader:    new(loaderMock.Loader),
		validator: new(engineMock.Validator),
		mutator:   new(engineMock.Mutator),
	}

	handler := asset.New(log, fakeRecorder(), mocks.store, mocks.loader, bucketStatusFinder, mocks.validator, mocks.mutator, relistInterval)

	return handler, mocks
}

func fakeRecorder() record.EventRecorder {
	return record.NewFakeRecorder(20)
}

func testData(assetName, bucketName, url string) *v1alpha2.Asset {
	return &v1alpha2.Asset{
		ObjectMeta: v1.ObjectMeta{
			Name:       assetName,
			Generation: int64(1),
		},
		Spec: v1alpha2.AssetSpec{
			CommonAssetSpec: v1alpha2.CommonAssetSpec{
				BucketRef: v1alpha2.AssetBucketRef{Name: bucketName},
				Source: v1alpha2.AssetSource{
					URL:                      url,
					Mode:                     v1alpha2.AssetSingle,
					ValidationWebhookService: make([]v1alpha2.AssetWebhookService, 3),
					MutationWebhookService:   make([]v1alpha2.AssetWebhookService, 3),
				},
			},
		},
	}
}
