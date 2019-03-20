package asset_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	automock3 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset/pretty"
	automock2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/loader/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store/automock"
	"github.com/onsi/gomega"
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

func TestAssetHandler_IsOnAddOrUpdate(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := new(v1alpha2.Asset)
		testData.ObjectMeta.Generation = int64(1)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnAddOrUpdate(testData, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Updated", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.ObjectMeta.Generation = int64(10)
		testData.Status.CommonAssetStatus.ObservedGeneration = int64(8)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnAddOrUpdate(testData, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("NotChanged", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.ObjectMeta.Generation = int64(10)
		testData.Status.CommonAssetStatus.ObservedGeneration = int64(10)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnAddOrUpdate(testData, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})
}

func TestAssetHandler_IsOnDelete(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		now := v1.Now()
		testData.ObjectMeta.DeletionTimestamp = &now
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnDelete(testData)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("False", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnDelete(testData)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})
}

func TestAssetHandler_IsOnFailed(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnFailed(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Ready", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnFailed(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("True", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetFailed
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnFailed(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func TestAssetHandler_IsOnPending(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnPending(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Ready", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnPending(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("True", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnPending(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func TestAssetHandler_IsOnReady(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		testData := new(v1alpha2.Asset)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnReady(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("Pending", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetPending
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnReady(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("True", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.IsOnReady(testData.Status.CommonAssetStatus)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})
}

func TestAssetHandler_OnAddOrUpdate(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnAddOrUpdate(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("SuccessUpdate", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.CommonAssetStatus.AssetRef.Assets = make([]string, 10)
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("DeleteObjects", ctx, testData.Spec.BucketRef.Name, fmt.Sprintf("/%s", testData.Name)).Return(nil).Once()
		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnAddOrUpdate(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("CleanupError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.CommonAssetStatus.AssetRef.Assets = make([]string, 10)
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("DeleteObjects", ctx, testData.Spec.BucketRef.Name, fmt.Sprintf("/%s", testData.Name)).Return(errors.New("test")).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnAddOrUpdate(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.CleanupError.String()))
	})
}

func TestAssetHandler_OnDelete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("DeleteObjects", ctx, "bucket-name", fmt.Sprintf("/%s", testData.Name)).Return(nil).Once()
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		err := assetHandler.OnDelete(ctx, testData, testData.Spec.CommonAssetSpec)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("BucketNotReady", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		testData.Spec.CommonAssetSpec.BucketRef.Name = "notReady"
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		err := assetHandler.OnDelete(ctx, testData, testData.Spec.CommonAssetSpec)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("BucketError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		testData.Spec.CommonAssetSpec.BucketRef.Name = "error"
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		err := assetHandler.OnDelete(ctx, testData, testData.Spec.CommonAssetSpec)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("DeleteError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("DeleteObjects", ctx, "bucket-name", fmt.Sprintf("/%s", testData.Name)).Return(errors.New("test")).Once()
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		err := assetHandler.OnDelete(ctx, testData, testData.Spec.CommonAssetSpec)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestAssetHandler_OnFailed(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status, err := assetHandler.OnFailed(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("CleanupError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.Reason = pretty.CleanupError.String()
		testData.Status.AssetRef.Assets = make([]string, 10)
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("DeleteObjects", ctx, testData.Spec.BucketRef.Name, fmt.Sprintf("/%s", testData.Name)).Return(nil).Once()
		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status, err := assetHandler.OnFailed(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("StillFailingWithSameReason", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.Reason = pretty.CleanupError.String()
		testData.Status.Phase = v1alpha2.AssetFailed
		testData.Status.AssetRef.Assets = make([]string, 10)
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("DeleteObjects", ctx, testData.Spec.BucketRef.Name, fmt.Sprintf("/%s", testData.Name)).Return(errors.New("test")).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		_, err := assetHandler.OnFailed(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestAssetHandler_OnPending(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("SuccessNoWebhooks", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Spec.CommonAssetSpec.Source.MutationWebhookService = nil
		testData.Spec.CommonAssetSpec.Source.ValidationWebhookService = nil
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(nil).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("LoadError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, errors.New("err")).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.PullingFailed.String()))
	})

	t.Run("MutationError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(errors.New("err")).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.MutationFailed.String()))
	})

	t.Run("ValidationFailed", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: false}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.ValidationFailed.String()))
	})

	t.Run("ValidationError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: false}, errors.New("test")).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.ValidationError.String()))
	})

	t.Run("UploadError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)
		loader := new(automock2.Loader)
		defer loader.AssertExpectations(t)
		mutator := new(automock3.Mutator)
		defer mutator.AssertExpectations(t)
		validator := new(automock3.Validator)
		defer validator.AssertExpectations(t)

		store.On("PutObjects", ctx, testData.Spec.BucketRef.Name, testData.Name, "/tmp", mock.AnythingOfType("[]string")).Return(errors.New("test")).Once()
		loader.On("Load", testData.Spec.Source.Url, testData.Name, testData.Spec.Source.Mode, testData.Spec.Source.Filter).Return("/tmp", nil, nil).Once()
		loader.On("Clean", "/tmp").Return(nil).Once()
		mutator.On("Mutate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.MutationWebhookService).Return(nil).Once()
		validator.On("Validate", ctx, testData, "/tmp", mock.AnythingOfType("[]string"), testData.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
		assetHandler := asset.New(fakeRecorder(), store, loader, bucketStatusFinder, validator, mutator, log)

		// When
		status := assetHandler.OnPending(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.UploadFailed.String()))
	})
}

func TestAssetHandler_OnReady(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("ContainsAllObjects", ctx, "bucket-name", testData.Name, mock.AnythingOfType("[]string")).Return(true, nil).Once()
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		status := assetHandler.OnReady(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	})

	t.Run("BucketNotReady", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		testData.Spec.CommonAssetSpec.BucketRef.Name = "notReady"
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		status := assetHandler.OnReady(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketNotReady.String()))
	})

	t.Run("BucketError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		testData.Spec.CommonAssetSpec.BucketRef.Name = "error"
		ctx := context.TODO()

		store := new(automock.Store)
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		status := assetHandler.OnReady(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketError.String()))
	})

	t.Run("ContainsError", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("ContainsAllObjects", ctx, "bucket-name", testData.Name, mock.AnythingOfType("[]string")).Return(false, errors.New("error")).Once()
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		status := assetHandler.OnReady(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.RemoteContentVerificationError.String()))
	})

	t.Run("Missing", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		// Given
		testData := new(v1alpha2.Asset)
		testData.Status.CommonAssetStatus.Phase = v1alpha2.AssetReady
		testData.ObjectMeta.Name = "test"
		ctx := context.TODO()

		store := new(automock.Store)
		store.On("ContainsAllObjects", ctx, "bucket-name", testData.Name, mock.AnythingOfType("[]string")).Return(false, nil).Once()
		defer store.AssertExpectations(t)

		assetHandler := asset.New(fakeRecorder(), store, nil, bucketStatusFinder, nil, nil, log)

		// When
		status := assetHandler.OnReady(ctx, testData, testData.Spec.CommonAssetSpec, testData.Status.CommonAssetStatus)

		// Then
		g.Expect(status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.MissingContent.String()))
	})
}

func TestAssetHandler_ShouldReconcile(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Updated", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.ObjectMeta.Generation = int64(2)
		testData.Status.ObservedGeneration = int64(1)
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("BeingDeleted", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		deletion := v1.Now()
		testData.ObjectMeta.DeletionTimestamp = &deletion
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("OnReady", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetReady
		testData.Status.LastHeartbeatTime = v1.NewTime(now.Add(-10 * relistInterval))
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("OnReadySkip", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetReady
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("OnPendingBucketNotReady", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetPending
		testData.Status.Reason = pretty.BucketNotReady.String()
		testData.Status.LastHeartbeatTime = v1.NewTime(now.Add(-10 * relistInterval))
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("OnPending", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetPending
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("OnPendingBucketNotReadySkip", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetPending
		testData.Status.Reason = pretty.BucketNotReady.String()
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("OnFailedValidationFail", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetFailed
		testData.Status.Reason = pretty.ValidationFailed.String()
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("OnFailedMutationFail", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetFailed
		testData.Status.Reason = pretty.MutationFailed.String()
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeFalse())
	})

	t.Run("OnFailedBucketError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		now := time.Now()
		relistInterval := time.Second

		testData := testData("test", "bucket-name", "https://test.com/file.txt")
		testData.Status.ObservedGeneration = testData.ObjectMeta.Generation
		testData.Status.Phase = v1alpha2.AssetFailed
		testData.Status.Reason = pretty.BucketError.String()
		testData.Status.LastHeartbeatTime = v1.NewTime(now)
		assetHandler := asset.New(nil, nil, nil, nil, nil, nil, log)

		// When
		result := assetHandler.ShouldReconcile(testData, testData.Status.CommonAssetStatus, now, relistInterval)

		// Then
		g.Expect(result).To(gomega.BeTrue())
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
			Url:        "http://test-url.com/bucket-name",
			RemoteName: "bucket-name",
		}, true, nil
	}
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
					Url:                      url,
					Mode:                     v1alpha2.AssetSingle,
					ValidationWebhookService: make([]v1alpha2.AssetWebhookService, 3),
					MutationWebhookService:   make([]v1alpha2.AssetWebhookService, 3),
				},
			},
		},
	}
}
