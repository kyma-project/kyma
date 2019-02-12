package asset

import (
	"context"
	"fmt"
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/engine"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/asset/pretty"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/loader"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Handler interface {
	ShouldReconcile(object MetaAccessor, status v1alpha1.CommonAssetStatus, now time.Time, relistInterval time.Duration) bool
	IsOnAddOrUpdate(object MetaAccessor, status v1alpha1.CommonAssetStatus) bool
	IsOnPending(status v1alpha1.CommonAssetStatus) bool
	IsOnDelete(object MetaAccessor) bool
	IsOnFailed(status v1alpha1.CommonAssetStatus) bool
	IsOnReady(status v1alpha1.CommonAssetStatus) bool
	OnAddOrUpdate(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus
	OnFailed(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) (*v1alpha1.CommonAssetStatus, error)
	OnReady(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus
	OnDelete(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec) error
	OnPending(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus
}

type MetaAccessor interface {
	GetNamespace() string
	GetName() string
	GetGeneration() int64
	GetDeletionTimestamp() *v1.Time
	GetFinalizers() []string
	SetFinalizers(finalizers []string)
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

var _ Handler = &assetHandler{}

type FindBucketStatus func(ctx context.Context, namespace, name string) (*v1alpha1.CommonBucketStatus, bool, error)

type assetHandler struct {
	recorder         record.EventRecorder
	findBucketStatus FindBucketStatus
	store            store.Store
	loader           loader.Loader
	validator        engine.Validator
	mutator          engine.Mutator
}

func New(recorder record.EventRecorder, store store.Store, loader loader.Loader, findBucketFnc FindBucketStatus, validator engine.Validator, mutator engine.Mutator) *assetHandler {
	return &assetHandler{
		recorder:         recorder,
		store:            store,
		loader:           loader,
		findBucketStatus: findBucketFnc,
		validator:        validator,
		mutator:          mutator,
	}
}

func (h *assetHandler) ShouldReconcile(object MetaAccessor, status v1alpha1.CommonAssetStatus, now time.Time, relistInterval time.Duration) bool {
	if h.IsOnDelete(object) {
		return true
	}

	if h.IsOnAddOrUpdate(object, status) {
		return true
	}

	if h.IsOnReady(status) && now.Before(status.LastHeartbeatTime.Add(relistInterval)) {
		return false
	}

	if h.IsOnPending(status) && status.Reason == pretty.BucketNotReady.String() && now.Before(status.LastHeartbeatTime.Add(relistInterval)) {
		return false
	}

	if h.IsOnFailed(status) && (status.Reason == pretty.ValidationFailed.String() || status.Reason == pretty.MutationFailed.String()) {
		return false
	}

	return true
}

func (*assetHandler) IsOnAddOrUpdate(object MetaAccessor, status v1alpha1.CommonAssetStatus) bool {
	return status.ObservedGeneration != object.GetGeneration()
}

func (*assetHandler) IsOnPending(status v1alpha1.CommonAssetStatus) bool {
	return status.Phase == v1alpha1.AssetPending
}

func (*assetHandler) IsOnDelete(object MetaAccessor) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func (*assetHandler) IsOnFailed(status v1alpha1.CommonAssetStatus) bool {
	return status.Phase == v1alpha1.AssetFailed
}

func (*assetHandler) IsOnReady(status v1alpha1.CommonAssetStatus) bool {
	return status.Phase == v1alpha1.AssetReady
}

func (h *assetHandler) OnAddOrUpdate(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
	if len(status.AssetRef.Assets) > 0 {
		if err := h.OnDelete(ctx, object, spec); err != nil {
			h.recordWarningEventf(object, pretty.CleanupError, err.Error())
			return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.CleanupError, err.Error()))
		}
		h.recordNormalEventf(object, pretty.Cleaned)
	}

	return h.OnPending(ctx, object, spec, status)
}

func (h *assetHandler) OnFailed(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) (*v1alpha1.CommonAssetStatus, error) {
	var newStatus v1alpha1.CommonAssetStatus
	switch status.Reason {
	case pretty.CleanupError.String():
		newStatus = h.OnAddOrUpdate(ctx, object, spec, status)
	default:
		newStatus = h.OnPending(ctx, object, spec, status)
	}

	if status.Reason == newStatus.Reason && status.Phase == newStatus.Phase {
		return nil, errors.New(status.Message)
	}

	return &newStatus, nil
}

func (h *assetHandler) OnDelete(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec) error {
	bucketStatus, isReady, err := h.findBucketStatus(ctx, object.GetNamespace(), spec.BucketRef.Name)
	if err != nil {
		return errors.Wrap(err, "while reading bucket status")
	}
	if !isReady {
		return nil
	}

	if err := h.store.DeleteObjects(ctx, bucketStatus.RemoteName, fmt.Sprintf("/%s", object.GetName())); err != nil {
		return errors.Wrap(err, "while deleting asset content")
	}

	return nil
}

func (h *assetHandler) OnReady(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
	bucketStatus, isReady, err := h.findBucketStatus(ctx, object.GetNamespace(), spec.BucketRef.Name)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketError, err.Error())
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.BucketError, err.Error()))
	}
	if !isReady {
		h.recordWarningEventf(object, pretty.BucketNotReady)
		return h.getStatus(object, status, v1alpha1.AssetPending, withReasonStatus(pretty.BucketNotReady))
	}

	exists, err := h.store.ContainsAllObjects(ctx, bucketStatus.RemoteName, object.GetName(), status.AssetRef.Assets)
	if err != nil {
		h.recordWarningEventf(object, pretty.RemoteContentVerificationError, err.Error())
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.RemoteContentVerificationError, err.Error()))
	}
	if !exists {
		h.recordWarningEventf(object, pretty.MissingContent)
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.MissingContent))
	}

	return h.getStatus(object, status, v1alpha1.AssetReady)
}

func (h *assetHandler) OnPending(ctx context.Context, object MetaAccessor, spec v1alpha1.CommonAssetSpec, status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
	bucketStatus, isReady, err := h.findBucketStatus(ctx, object.GetNamespace(), spec.BucketRef.Name)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketError, err.Error())
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.BucketError, err.Error()))
	}
	if !isReady {
		h.recordWarningEventf(object, pretty.BucketNotReady)
		return h.getStatus(object, status, v1alpha1.AssetPending, withReasonStatus(pretty.BucketNotReady))
	}

	basePath, files, err := h.loader.Load(spec.Source.Url, object.GetName(), spec.Source.Mode, spec.Source.Filter)
	defer h.loader.Clean(basePath)
	if err != nil {
		h.recordWarningEventf(object, pretty.PullingFailed, err.Error())
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.PullingFailed, err.Error()))
	}
	h.recordNormalEventf(object, pretty.Pulled)

	if len(spec.Source.MutationWebhookService) > 0 {
		if err := h.mutator.Mutate(ctx, object, basePath, files, spec.Source.MutationWebhookService); err != nil {
			h.recordWarningEventf(object, pretty.MutationFailed, err.Error())
			return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.MutationFailed, err.Error()))
		}
		h.recordNormalEventf(object, pretty.Mutated)
	}

	if len(spec.Source.ValidationWebhookService) > 0 {
		result, err := h.validator.Validate(ctx, object, basePath, files, spec.Source.ValidationWebhookService)
		if err != nil {
			h.recordWarningEventf(object, pretty.ValidationError, err.Error())
			return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.ValidationError, err.Error()))
		}
		if !result.Success {
			h.recordWarningEventf(object, pretty.ValidationFailed, result.Messages)
			return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.ValidationFailed, result.Messages))
		}
		h.recordNormalEventf(object, pretty.Validated)
	}

	if err := h.store.PutObjects(ctx, bucketStatus.RemoteName, object.GetName(), basePath, files); err != nil {
		h.recordWarningEventf(object, pretty.UploadFailed, err.Error())
		return h.getStatus(object, status, v1alpha1.AssetFailed, withReasonStatus(pretty.UploadFailed, err.Error()))
	}
	h.recordNormalEventf(object, pretty.Uploaded)

	return h.getStatus(object, status, v1alpha1.AssetReady, withReasonStatus(pretty.Uploaded), withFilesStatus(h.getBaseUrl(bucketStatus.Url, object.GetName()), files))
}

func (h *assetHandler) getBaseUrl(bucketUrl, assetName string) string {
	return fmt.Sprintf("%s/%s", bucketUrl, assetName)
}

func (h *assetHandler) recordNormalEventf(object MetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Normal", reason, args...)
}

func (h *assetHandler) recordWarningEventf(object MetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Warning", reason, args...)
}

func (h *assetHandler) recordEventf(object MetaAccessor, eventType string, reason pretty.Reason, args ...interface{}) {
	h.recorder.Eventf(object, eventType, reason.String(), reason.Message(), args...)
}

type StatusOption func(v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus

func (*assetHandler) getStatus(object MetaAccessor, status v1alpha1.CommonAssetStatus, phase v1alpha1.AssetPhase, options ...StatusOption) v1alpha1.CommonAssetStatus {
	status.LastHeartbeatTime = v1.Now()
	status.ObservedGeneration = object.GetGeneration()
	status.Phase = phase

	for _, option := range options {
		status = option(status)
	}

	return status
}

func withReasonStatus(reason pretty.Reason, args ...interface{}) func(v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
	return func(status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
		status.Reason = reason.String()
		if len(reason.Message()) > 0 {
			status.Message = fmt.Sprintf(reason.Message(), args...)
		}
		return status
	}
}

func withFilesStatus(baseUrl string, files []string) func(v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
	return func(status v1alpha1.CommonAssetStatus) v1alpha1.CommonAssetStatus {
		status.AssetRef.BaseUrl = baseUrl
		status.AssetRef.Assets = files

		return status
	}
}
