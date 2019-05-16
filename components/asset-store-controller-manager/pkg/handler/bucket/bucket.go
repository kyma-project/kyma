package bucket

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/bucket/pretty"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
)

type Handler interface {
	Do(ctx context.Context, now time.Time, instance MetaAccessor, spec v1alpha2.CommonBucketSpec, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error)
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
	log              logr.Logger
	relistInterval   time.Duration
}

func New(log logr.Logger, recorder record.EventRecorder, store store.Store, externalEndpoint string, relistInterval time.Duration) Handler {
	return &bucketHandler{
		recorder:         recorder,
		store:            store,
		externalEndpoint: externalEndpoint,
		log:              log,
		relistInterval:   relistInterval,
	}
}

func (h *bucketHandler) Do(ctx context.Context, now time.Time, instance MetaAccessor, spec v1alpha2.CommonBucketSpec, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error) {
	h.logInfof("Start common Bucket handling")
	defer h.logInfof("Finish common Bucket handling")

	switch {
	case h.isOnDelete(instance):
		return h.onDelete(ctx, instance, status)
	case h.isOnAddOrUpdate(instance, status):
		return h.onAddOrUpdate(instance, spec, status)
	case h.isOnReady(status, now):
		return h.onReady(instance, spec, status)
	case h.isOnFailed(status):
		return h.onFailed(instance, spec, status)
	default:
		h.logInfof("Action not taken")
		return nil, nil
	}
}

func (h *bucketHandler) isOnReady(status v1alpha2.CommonBucketStatus, now time.Time) bool {
	return status.Phase == v1alpha2.BucketReady && now.After(status.LastHeartbeatTime.Add(h.relistInterval))
}

func (*bucketHandler) isOnAddOrUpdate(object MetaAccessor, status v1alpha2.CommonBucketStatus) bool {
	return status.ObservedGeneration != object.GetGeneration()
}

func (*bucketHandler) isOnFailed(status v1alpha2.CommonBucketStatus) bool {
	return status.Phase == v1alpha2.BucketFailed
}

func (*bucketHandler) isOnDelete(object MetaAccessor) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func (h *bucketHandler) onFailed(object MetaAccessor, spec v1alpha2.CommonBucketSpec, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error) {
	switch status.Reason {
	case pretty.BucketNotFound.String():
		return h.onAddOrUpdate(object, spec, status)
	case pretty.BucketCreationFailure.String():
		return h.onAddOrUpdate(object, spec, status)
	case pretty.BucketVerificationFailure.String():
		return h.onReady(object, spec, status)
	case pretty.BucketPolicyUpdateFailed.String():
		return h.onReady(object, spec, status)
	}

	return nil, nil
}

func (h *bucketHandler) onReady(object MetaAccessor, spec v1alpha2.CommonBucketSpec, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error) {
	h.logInfof("Checking if bucket exists")
	exists, err := h.store.BucketExists(status.RemoteName)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketVerificationFailure, err.Error())
		return h.getStatus(object, status.RemoteName, status.URL, v1alpha2.BucketFailed, pretty.BucketVerificationFailure, err.Error()), err
	}
	if !exists {
		h.recordWarningEventf(object, pretty.BucketNotFound, status.RemoteName)
		return h.getStatus(object, "", "", v1alpha2.BucketFailed, pretty.BucketNotFound, status.RemoteName), errors.Errorf(pretty.BucketNotFound.String(), status.RemoteName)
	}
	h.logInfof("Bucket exists")

	h.logInfof("Comparing bucket policy")
	equal, err := h.store.CompareBucketPolicy(status.RemoteName, spec.Policy)
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketPolicyVerificationFailed, err.Error())
		return h.getStatus(object, status.RemoteName, status.URL, v1alpha2.BucketFailed, pretty.BucketPolicyVerificationFailed, status.RemoteName), err
	}
	if equal {
		h.logInfof("Bucket is up-to-date")
		return h.getStatus(object, status.RemoteName, status.URL, v1alpha2.BucketReady, pretty.BucketPolicyUpdated), nil
	}

	h.logInfof("Updating bucket policy")
	h.recordWarningEventf(object, pretty.BucketPolicyHasBeenChanged)
	if err := h.store.SetBucketPolicy(status.RemoteName, spec.Policy); err != nil {
		h.recordWarningEventf(object, pretty.BucketPolicyUpdateFailed, err.Error())
		return h.getStatus(object, status.RemoteName, status.URL, v1alpha2.BucketFailed, pretty.BucketPolicyUpdateFailed, err.Error()), err
	}
	h.recordNormalEventf(object, pretty.BucketPolicyUpdated)
	h.logInfof("Bucket policy updated")

	return h.getStatus(object, status.RemoteName, status.URL, v1alpha2.BucketReady, pretty.BucketPolicyUpdated), nil
}

func (h *bucketHandler) onAddOrUpdate(object MetaAccessor, spec v1alpha2.CommonBucketSpec, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error) {
	h.logInfof("Checking if bucket was previously created")
	if status.RemoteName != "" {
		h.logInfof("Bucket was created")
		return h.onReady(object, spec, status)
	}

	h.logInfof("Creating bucket")
	remoteName, err := h.store.CreateBucket(object.GetNamespace(), object.GetName(), string(spec.Region))
	if err != nil {
		h.recordWarningEventf(object, pretty.BucketCreationFailure, err.Error())
		return h.getStatus(object, "", "", v1alpha2.BucketFailed, pretty.BucketCreationFailure, err.Error()), err
	}
	h.recordNormalEventf(object, pretty.BucketCreated)
	h.logInfof("Bucket created")

	externalUrl := h.getBucketUrl(remoteName)

	h.logInfof("Updating bucket policy")
	if err := h.store.SetBucketPolicy(remoteName, spec.Policy); err != nil {
		h.recordWarningEventf(object, pretty.BucketPolicyUpdateFailed, err.Error())
		return h.getStatus(object, remoteName, externalUrl, v1alpha2.BucketFailed, pretty.BucketPolicyUpdateFailed, err.Error()), err
	}
	h.recordNormalEventf(object, pretty.BucketPolicyUpdated)
	h.logInfof("Bucket policy updated")

	return h.getStatus(object, remoteName, externalUrl, v1alpha2.BucketReady, pretty.BucketPolicyUpdated), nil
}

func (h *bucketHandler) onDelete(ctx context.Context, object MetaAccessor, status v1alpha2.CommonBucketStatus) (*v1alpha2.CommonBucketStatus, error) {
	h.logInfof("Deleting Bucket")
	if status.RemoteName == "" || status.Reason == pretty.BucketNotFound.String() {
		h.logInfof("Nothing to delete, there is no remote bucket")
		return nil, nil
	}

	if err := h.store.DeleteBucket(ctx, status.RemoteName); err != nil {
		return nil, errors.Wrap(err, "while deleting remote bucket")
	}
	h.logInfof("Remote bucket %s deleted", status.RemoteName)

	return nil, nil
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

func (h *bucketHandler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}

func (h *bucketHandler) recordEventf(object MetaAccessor, eventType string, reason pretty.Reason, args ...interface{}) {
	h.recorder.Eventf(object, eventType, reason.String(), reason.Message(), args...)
}

func (*bucketHandler) getStatus(object MetaAccessor, remoteName, url string, phase v1alpha2.BucketPhase, reason pretty.Reason, args ...interface{}) *v1alpha2.CommonBucketStatus {
	return &v1alpha2.CommonBucketStatus{
		LastHeartbeatTime:  v1.Now(),
		ObservedGeneration: object.GetGeneration(),
		Phase:              phase,
		RemoteName:         remoteName,
		URL:                url,
		Reason:             reason.String(),
		Message:            fmt.Sprintf(reason.Message(), args...),
	}
}
