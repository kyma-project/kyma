package docstopic

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic/pretty"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type CommonAsset struct {
	v1.ObjectMeta
	Spec   v1alpha2.CommonAssetSpec
	Status v1alpha2.CommonAssetStatus
}

const (
	docsTopicLabel = "docstopic.cms.kyma-project.io"
	typeLabel      = "type.cms.kyma-project.io"
	accessLabel    = "access.cms.kyma-project.io"
)

//go:generate mockery -name=AssetService -output=automock -outpkg=automock -case=underscore
type AssetService interface {
	List(ctx context.Context, namespace string, labels map[string]string) ([]CommonAsset, error)
	Create(ctx context.Context, docsTopic v1.Object, commonAsset CommonAsset) error
	Update(ctx context.Context, commonAsset CommonAsset) error
	Delete(ctx context.Context, commonAsset CommonAsset) error
}

//go:generate mockery -name=BucketService -output=automock -outpkg=automock -case=underscore
type BucketService interface {
	List(ctx context.Context, namespace string, labels map[string]string) ([]string, error)
	Create(ctx context.Context, namespacedName types.NamespacedName, private bool, labels map[string]string) error
}

type Handler interface {
	Handle(ctx context.Context, instance ObjectMetaAccessor, spec v1alpha1.CommonDocsTopicSpec, status v1alpha1.CommonDocsTopicStatus) (*v1alpha1.CommonDocsTopicStatus, error)
}

type ObjectMetaAccessor interface {
	v1.Object
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

type docstopicHandler struct {
	log       logr.Logger
	recorder  record.EventRecorder
	assetSvc  AssetService
	bucketSvc BucketService
}

func New(log logr.Logger, recorder record.EventRecorder, assetSvc AssetService, bucketSvc BucketService) Handler {
	return &docstopicHandler{
		log:       log,
		recorder:  recorder,
		assetSvc:  assetSvc,
		bucketSvc: bucketSvc,
	}
}

func (h *docstopicHandler) Handle(ctx context.Context, instance ObjectMetaAccessor, spec v1alpha1.CommonDocsTopicSpec, status v1alpha1.CommonDocsTopicStatus) (*v1alpha1.CommonDocsTopicStatus, error) {
	h.logInfof("Start common DocsTopic handling")
	defer h.logInfof("Finish common DocsTopic handling")

	bucketName, err := h.ensureBucketExits(ctx, instance.GetNamespace())
	if err != nil {
		h.recordWarningEventf(instance, pretty.BucketError, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, pretty.BucketError, err.Error()), status), err
	}

	commonAssets, err := h.assetSvc.List(ctx, instance.GetNamespace(), h.buildLabel(instance.GetName(), ""))
	if err != nil {
		h.recordWarningEventf(instance, pretty.AssetsListingFailed, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, pretty.AssetsListingFailed, err.Error()), status), err
	}
	commonAssetsMap := h.convertToAssetMap(commonAssets)

	switch {
	case h.isOnChange(commonAssetsMap, spec, bucketName):
		return h.onChange(ctx, instance, spec, status, commonAssetsMap, bucketName)
	case h.isOnPhaseChange(commonAssetsMap, status):
		return h.onPhaseChange(instance, status, commonAssetsMap)
	default:
		h.logInfof("Instance is up-to-date, action not taken")
		return nil, nil
	}
}

func (h *docstopicHandler) ensureBucketExits(ctx context.Context, namespace string) (string, error) {
	h.logInfof("Listing buckets")
	labels := map[string]string{accessLabel: "public"}
	names, err := h.bucketSvc.List(ctx, namespace, labels)
	if err != nil {
		return "", err
	}

	bucketCount := len(names)
	if bucketCount > 1 {
		return "", fmt.Errorf("too many buckets with labels: %+v", labels)
	}
	if bucketCount == 1 {
		h.logInfof("Bucket %s already exits", names[0])
		return names[0], nil
	}

	name := h.generateBucketName(false)
	h.logInfof("Creating bucket %s", name)
	if err := h.bucketSvc.Create(ctx, types.NamespacedName{Name: name, Namespace: namespace}, false, labels); err != nil {
		return "", err
	}
	h.logInfof("Bucket created %s", name)

	return name, nil
}

func (h *docstopicHandler) isOnChange(existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string) bool {
	return h.shouldCreateAssets(existing, spec) || h.shouldDeleteAssets(existing, spec) || h.shouldUpdateAssets(existing, spec, bucketName)
}

func (h *docstopicHandler) isOnPhaseChange(existing map[string]CommonAsset, status v1alpha1.CommonDocsTopicStatus) bool {
	return status.Phase != h.calculateAssetPhase(existing)
}

func (h *docstopicHandler) shouldCreateAssets(existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) bool {
	for key := range spec.Sources {
		if _, exists := existing[key]; !exists {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) shouldUpdateAssets(existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string) bool {
	for key, existingAsset := range existing {
		expectedSpec, exists := spec.Sources[key]
		if !exists {
			continue
		}

		expected := h.convertToCommonAssetSpec(expectedSpec, bucketName)
		if !reflect.DeepEqual(expected, existingAsset.Spec) {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) shouldDeleteAssets(existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) bool {
	for key := range existing {
		if _, exists := spec.Sources[key]; !exists {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) onPhaseChange(instance ObjectMetaAccessor, status v1alpha1.CommonDocsTopicStatus, existing map[string]CommonAsset) (*v1alpha1.CommonDocsTopicStatus, error) {
	phase := h.calculateAssetPhase(existing)
	h.logInfof("Updating phase to %s", phase)

	if phase == v1alpha1.DocsTopicPending {
		h.recordNormalEventf(instance, pretty.WaitingForAssets)
		return h.buildStatus(phase, pretty.WaitingForAssets), nil
	}

	h.recordNormalEventf(instance, pretty.AssetsReady)
	return h.buildStatus(phase, pretty.AssetsReady), nil
}

func (h *docstopicHandler) onChange(ctx context.Context, instance ObjectMetaAccessor, spec v1alpha1.CommonDocsTopicSpec, status v1alpha1.CommonDocsTopicStatus, existing map[string]CommonAsset, bucketName string) (*v1alpha1.CommonDocsTopicStatus, error) {
	if err := h.createMissingAssets(ctx, instance, existing, spec, bucketName); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, pretty.AssetsCreationFailed, err.Error()), status), err
	}

	if err := h.updateOutdatedAssets(ctx, instance, existing, spec, bucketName); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, pretty.AssetsUpdateFailed, err.Error()), status), err
	}

	if err := h.deleteNotExisting(ctx, instance, existing, spec); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, pretty.AssetsDeletionFailed, err.Error()), status), err
	}

	h.recordNormalEventf(instance, pretty.WaitingForAssets)
	return h.buildStatus(v1alpha1.DocsTopicPending, pretty.WaitingForAssets), nil
}

func (h *docstopicHandler) createMissingAssets(ctx context.Context, instance ObjectMetaAccessor, existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string) error {
	for key, spec := range spec.Sources {
		if _, exists := existing[key]; exists {
			continue
		}

		if err := h.createAsset(ctx, instance, key, spec, bucketName); err != nil {
			return err
		}
	}

	return nil
}

func (h *docstopicHandler) createAsset(ctx context.Context, instance ObjectMetaAccessor, assetType string, assetSpec v1alpha1.Source, bucketName string) error {
	commonAsset := CommonAsset{
		ObjectMeta: v1.ObjectMeta{
			Name:      h.generateAssetName(instance.GetName(), assetType),
			Namespace: instance.GetNamespace(),
			Labels:    h.buildLabel(instance.GetName(), assetType),
		},
		Spec: h.convertToCommonAssetSpec(assetSpec, bucketName),
	}

	h.logInfof("Creating asset %s", commonAsset.Name)
	if err := h.assetSvc.Create(ctx, instance, commonAsset); err != nil {
		h.recordWarningEventf(instance, pretty.AssetCreationFailed, commonAsset.Name, err.Error())
		return err
	}
	h.logInfof("Asset %s created", commonAsset.Name)
	h.recordNormalEventf(instance, pretty.AssetCreated, commonAsset.Name)

	return nil
}

func (h *docstopicHandler) updateOutdatedAssets(ctx context.Context, instance ObjectMetaAccessor, existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string) error {
	for key, existingAsset := range existing {
		expectedSpec, exists := spec.Sources[key]
		if !exists {
			continue
		}

		h.logInfof("Updating asset %s", existingAsset.Name)
		expected := h.convertToCommonAssetSpec(expectedSpec, bucketName)
		if reflect.DeepEqual(expected, existingAsset.Spec) {
			h.logInfof("Asset %s is up-to-date", existingAsset.Name)
			continue
		}

		existingAsset.Spec = expected
		if err := h.assetSvc.Update(ctx, existingAsset); err != nil {
			h.recordWarningEventf(instance, pretty.AssetUpdateFailed, existingAsset.Name, err.Error())
			return err
		}
		h.logInfof("Asset %s updated", existingAsset.Name)
		h.recordNormalEventf(instance, pretty.AssetUpdated, existingAsset.Name)
	}

	return nil
}

func (h *docstopicHandler) deleteNotExisting(ctx context.Context, instance ObjectMetaAccessor, existing map[string]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) error {
	for key, commonAsset := range existing {
		if _, exists := spec.Sources[key]; exists {
			continue
		}

		h.logInfof("Deleting asset %s", commonAsset.Name)
		if err := h.assetSvc.Delete(ctx, commonAsset); err != nil {
			h.recordWarningEventf(instance, pretty.AssetDeletionFailed, commonAsset.Name, err.Error())
			return err
		}
		h.logInfof("Asset %s deleted", commonAsset.Name)
		h.recordNormalEventf(instance, pretty.AssetDeleted, commonAsset.Name)
	}

	return nil
}

func (h *docstopicHandler) convertToAssetMap(assets []CommonAsset) map[string]CommonAsset {
	result := make(map[string]CommonAsset)

	for _, asset := range assets {
		sourceType := asset.Labels[typeLabel]
		result[sourceType] = asset
	}

	return result
}

func (h *docstopicHandler) convertToCommonAssetSpec(spec v1alpha1.Source, bucketName string) v1alpha2.CommonAssetSpec {
	return v1alpha2.CommonAssetSpec{
		Source: v1alpha2.AssetSource{
			Mode:   h.convertToAssetMode(spec.Mode),
			Url:    spec.URL,
			Filter: spec.Filter,
		},
		BucketRef: v1alpha2.AssetBucketRef{
			Name: bucketName,
		},
	}
}

func (h *docstopicHandler) buildLabel(topicName, assetType string) map[string]string {
	labels := make(map[string]string)

	labels[docsTopicLabel] = topicName
	if len(assetType) > 0 {
		labels[typeLabel] = assetType
	}

	return labels
}

func (h *docstopicHandler) convertToAssetMode(mode v1alpha1.DocsTopicMode) v1alpha2.AssetMode {
	switch mode {
	case v1alpha1.DocsTopicIndex:
		return v1alpha2.AssetIndex
	case v1alpha1.DocsTopicPackage:
		return v1alpha2.AssetPackage
	default:
		return v1alpha2.AssetSingle
	}
}

func (h *docstopicHandler) calculateAssetPhase(existing map[string]CommonAsset) v1alpha1.DocsTopicPhase {
	for _, asset := range existing {
		if asset.Status.Phase != v1alpha2.AssetReady {
			return v1alpha1.DocsTopicPending
		}
	}

	return v1alpha1.DocsTopicReady
}

func (h *docstopicHandler) buildStatus(phase v1alpha1.DocsTopicPhase, reason pretty.Reason, args ...interface{}) *v1alpha1.CommonDocsTopicStatus {
	return &v1alpha1.CommonDocsTopicStatus{
		Phase:             phase,
		Reason:            reason.String(),
		Message:           fmt.Sprintf(reason.Message(), args...),
		LastHeartbeatTime: v1.Now(),
	}
}

func (h *docstopicHandler) onFailedStatus(newStatus *v1alpha1.CommonDocsTopicStatus, oldStatus v1alpha1.CommonDocsTopicStatus) *v1alpha1.CommonDocsTopicStatus {
	if newStatus.Phase == oldStatus.Phase && newStatus.Reason == oldStatus.Reason {
		return nil
	}

	return newStatus
}

func (h *docstopicHandler) logInfof(message string, args ...interface{}) {
	h.log.Info(fmt.Sprintf(message, args...))
}

func (h *docstopicHandler) recordNormalEventf(object ObjectMetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Normal", reason, args...)
}

func (h *docstopicHandler) recordWarningEventf(object ObjectMetaAccessor, reason pretty.Reason, args ...interface{}) {
	h.recordEventf(object, "Warning", reason, args...)
}

func (h *docstopicHandler) recordEventf(object ObjectMetaAccessor, eventType string, reason pretty.Reason, args ...interface{}) {
	h.recorder.Eventf(object, eventType, reason.String(), reason.Message(), args...)
}

func (h *docstopicHandler) generateAssetName(docsTopicName, assetType string) string {
	assetTypeLower := strings.ToLower(assetType)
	return h.appendSuffix(fmt.Sprintf("%s-%s", docsTopicName, assetTypeLower))
}

func (h *docstopicHandler) generateBucketName(private bool) string {
	access := "public"
	if private {
		access = "private"
	}

	return h.appendSuffix(fmt.Sprintf("cms-%s", access))
}

func (h *docstopicHandler) appendSuffix(name string) string {
	unixNano := time.Now().UnixNano()
	suffix := strconv.FormatInt(unixNano, 32)

	return fmt.Sprintf("%s-%s", name, suffix)
}
