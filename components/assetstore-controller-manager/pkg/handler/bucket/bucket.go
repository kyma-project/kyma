package bucket

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/bucket/pretty"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"time"
)

type Handler interface {
	ShouldReconcile(object MetaAccessor, status v1alpha1.CommonBucketStatus, now time.Time, relistInterval time.Duration) bool
	IsOnAddOrUpdate(object MetaAccessor, status v1alpha1.CommonBucketStatus) bool
	IsOnDelete(object MetaAccessor) bool
	IsOnReady(status v1alpha1.CommonBucketStatus) bool
	IsOnFailed(status v1alpha1.CommonBucketStatus) bool
	OnAddOrUpdate(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus
	OnDelete(ctx context.Context, status v1alpha1.CommonBucketStatus) error
	OnReady(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus
	OnFailed(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) (*v1alpha1.CommonBucketStatus, error)
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

var _ Handler = &bucketHandler{}

type bucketHandler struct {
	recorder         record.EventRecorder
	store            store.Store
	externalEndpoint string
}

func New(recorder record.EventRecorder, store store.Store, externalEndpoint string) *bucketHandler {
	return &bucketHandler{
		recorder:         recorder,
		store:            store,
		externalEndpoint: externalEndpoint,
	}
}

func (h *bucketHandler) ShouldReconcile(object MetaAccessor, status v1alpha1.CommonBucketStatus, now time.Time, relistInterval time.Duration) bool {
	if h.IsOnDelete(object) {
		return true
	}

	if h.IsOnAddOrUpdate(object, status) {
		return true
	}

	if h.IsOnReady(status) && now.Before(status.LastHeartbeatTime.Add(relistInterval)) {
		return false
	}

	return true
}

func (*bucketHandler) IsOnReady(status v1alpha1.CommonBucketStatus) bool {
	return status.Phase == v1alpha1.BucketReady
}

func (*bucketHandler) IsOnAddOrUpdate(object MetaAccessor, status v1alpha1.CommonBucketStatus) bool {
	return status.ObservedGeneration != object.GetGeneration()
}

func (*bucketHandler) IsOnFailed(status v1alpha1.CommonBucketStatus) bool {
	return status.Phase == v1alpha1.BucketFailed && status.Reason != pretty.NotFoundReason.String()
}

func (*bucketHandler) IsOnDelete(object MetaAccessor) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func (h *bucketHandler) OnFailed(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) (*v1alpha1.CommonBucketStatus, error) {
	newStatus := status

	switch status.Reason {
	case pretty.BucketCreationFailure.String():
		newStatus = h.OnAddOrUpdate(object, spec, status)
	case pretty.BucketVerificationFailure.String():
		newStatus = h.OnReady(object, spec, status)
	case pretty.BucketPolicyUpdateFailed.String():
		newStatus = h.OnReady(object, spec, status)
	}

	if status.Phase == newStatus.Phase && status.Reason == newStatus.Reason {
		return nil, errors.New(status.Message)
	}

	return &newStatus, nil
}

func (h *bucketHandler) OnReady(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
	exists, err := h.store.BucketExists(status.RemoteName)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketVerificationFailure, err.Error())
		return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.BucketVerificationFailure, err.Error()))
	}
	if !exists {
		h.recordWarningEventf(object, pretty.NotFoundReason, status.RemoteName)
		return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.NotFoundReason, status.RemoteName))
	}

	equal, err := h.store.CompareBucketPolicy(status.RemoteName, spec.Policy)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketPolicyVerificationFailed, err.Error())
		return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.BucketPolicyVerificationFailed, status.RemoteName))
	}
	if !equal {
		h.recordWarningEventf(object, pretty.BucketPolicyHasBeenChanged)
		if err := h.store.SetBucketPolicy(status.RemoteName, spec.Policy); err != nil {
			h.recordWarningEventf(object, pretty.BucketPolicyUpdateFailed, err.Error())
			return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.BucketPolicyUpdateFailed, err.Error()), withBucketNameStatus(status.RemoteName), withUrlStatus(status.Url))
		}
		h.recordNormalEventf(object, pretty.BucketPolicyUpdated)
	}

	return h.getStatus(object, status, v1alpha1.BucketReady, withReasonStatus(pretty.BucketPolicyUpdated))
}

func (h *bucketHandler) OnAddOrUpdate(object MetaAccessor, spec v1alpha1.CommonBucketSpec, status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
	bucketName := status.RemoteName
	if len(bucketName) > 0 {
		return h.OnReady(object, spec, status)
	}

	bucketName, err := h.store.CreateBucket(object.GetNamespace(), object.GetName(), string(spec.Region))
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketCreationFailure, err.Error())
		return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.BucketCreationFailure, err.Error()))
	}
	h.recordNormalEventf(object, pretty.BucketCreated)

	externalUrl := h.getBucketUrl(bucketName)
	if err := h.store.SetBucketPolicy(bucketName, spec.Policy); err != nil {
		h.recordWarningEventf(object, pretty.BucketPolicyUpdateFailed, err.Error())
		return h.getStatus(object, status, v1alpha1.BucketFailed, withReasonStatus(pretty.BucketPolicyUpdateFailed, err.Error()), withBucketNameStatus(bucketName), withUrlStatus(externalUrl))
	}
	h.recordNormalEventf(object, pretty.BucketPolicyUpdated)

	return h.getStatus(object, status, v1alpha1.BucketReady, withReasonStatus(pretty.BucketPolicyUpdated), withBucketNameStatus(bucketName), withUrlStatus(externalUrl))
}

func (h *bucketHandler) OnDelete(ctx context.Context, status v1alpha1.CommonBucketStatus) error {
	if len(status.RemoteName) == 0 {
		return nil
	}

	if err := h.store.DeleteBucket(ctx, status.RemoteName); err != nil {
		return err
	}

	return nil
}

func (h *bucketHandler) getBucketUrl(name string) string {
	return fmt.Sprintf("%s/%s", h.externalEndpoint, name)
}

func (h *bucketHandler) recordNormalEventf(object MetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Normal", reason, args...)
}

func (h *bucketHandler) recordWarningEventf(object MetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Warning", reason, args...)
}

func (h *bucketHandler) recordEventf(object MetaAccessor, eventType string, reason pretty.Reason, args ...interface{}) {
	h.recorder.Eventf(object, eventType, reason.String(), reason.Message(), args...)
}

type statusOption func(v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus

func (*bucketHandler) getStatus(object MetaAccessor, status v1alpha1.CommonBucketStatus, phase v1alpha1.BucketPhase, options ...statusOption) v1alpha1.CommonBucketStatus {
	status.LastHeartbeatTime = v1.Now()
	status.ObservedGeneration = object.GetGeneration()
	status.Phase = phase

	for _, option := range options {
		status = option(status)
	}

	return status
}

func withUrlStatus(url string) func(v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
	return func(status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
		status.Url = url
		return status
	}
}

func withReasonStatus(reason pretty.Reason, args ...interface{}) func(v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
	return func(status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
		status.Reason = reason.String()
		if len(reason.Message()) > 0 {
			status.Message = fmt.Sprintf(reason.Message(), args...)
		}

		return status
	}
}

func withBucketNameStatus(bucketName string) func(v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
	return func(status v1alpha1.CommonBucketStatus) v1alpha1.CommonBucketStatus {
		status.RemoteName = bucketName
		return status
	}
}
