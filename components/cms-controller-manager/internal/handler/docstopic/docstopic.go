package docstopic

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

type CommonAsset struct {
	v1.ObjectMeta
	Spec   v1alpha2.CommonAssetSpec
	Status v1alpha2.CommonAssetStatus
}

const (
	docsTopicLabel           = "cms.kyma-project.io/docs-topic"
	accessLabel              = "cms.kyma-project.io/access"
	assetShortNameAnnotation = "cms.kyma-project.io/asset-short-name"
	typeLabel                = "cms.kyma-project.io/type"
)

var (
	errDuplicatedAssetName = errors.New("duplicated asset name")
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
	log              logr.Logger
	recorder         record.EventRecorder
	assetSvc         AssetService
	bucketSvc        BucketService
	webhookConfigSvc webhookconfig.AssetWebhookConfigService
}

func New(log logr.Logger, recorder record.EventRecorder, assetSvc AssetService, bucketSvc BucketService, webhookConfigSvc webhookconfig.AssetWebhookConfigService) Handler {
	return &docstopicHandler{
		log:              log,
		recorder:         recorder,
		assetSvc:         assetSvc,
		bucketSvc:        bucketSvc,
		webhookConfigSvc: webhookConfigSvc,
	}
}

func (h *docstopicHandler) Handle(ctx context.Context, instance ObjectMetaAccessor, spec v1alpha1.CommonDocsTopicSpec, status v1alpha1.CommonDocsTopicStatus) (*v1alpha1.CommonDocsTopicStatus, error) {
	h.logInfof("Start common DocsTopic handling")
	defer h.logInfof("Finish common DocsTopic handling")

	err := h.validateSpec(spec)
	if err != nil {
		h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetsSpecValidationFailed, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsSpecValidationFailed, err.Error()), status), err
	}

	bucketName, err := h.ensureBucketExits(ctx, instance.GetNamespace())
	if err != nil {
		h.recordWarningEventf(instance, v1alpha1.DocsTopicBucketError, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicBucketError, err.Error()), status), err
	}

	commonAssets, err := h.assetSvc.List(ctx, instance.GetNamespace(), h.buildLabels(instance.GetName(), ""))
	if err != nil {
		h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetsListingFailed, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsListingFailed, err.Error()), status), err
	}

	commonAssetsMap := h.convertToAssetMap(commonAssets)

	webhookCfg, err := h.webhookConfigSvc.Get(ctx)
	if err != nil {
		h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetsWebhookGetFailed, err.Error())
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsWebhookGetFailed, err.Error()), status), err
	}
	h.logInfof("Webhook configuration loaded")

	switch {
	case h.isOnChange(commonAssetsMap, spec, bucketName, webhookCfg):
		return h.onChange(ctx, instance, spec, status, commonAssetsMap, bucketName, webhookCfg)
	case h.isOnPhaseChange(commonAssetsMap, status):
		return h.onPhaseChange(instance, status, commonAssetsMap)
	default:
		h.logInfof("Instance is up-to-date, action not taken")
		return nil, nil
	}
}

func (h *docstopicHandler) validateSpec(spec v1alpha1.CommonDocsTopicSpec) error {
	h.logInfof("validating CommonDocsTopicSpec")
	names := map[v1alpha1.DocsTopicSourceName]map[v1alpha1.DocsTopicSourceType]struct{}{}
	for _, src := range spec.Sources {
		if nameTypes, exists := names[src.Name]; exists {
			if _, exists := nameTypes[src.Type]; exists {
				return errDuplicatedAssetName
			}
			names[src.Name][src.Type] = struct{}{}
			continue
		}
		names[src.Name] = map[v1alpha1.DocsTopicSourceType]struct{}{}
		names[src.Name][src.Type] = struct{}{}
	}
	h.logInfof("CommonDocsTopicSpec validated")

	return nil
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

func (h *docstopicHandler) isOnChange(existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string, config webhookconfig.AssetWebhookConfigMap) bool {
	return h.shouldCreateAssets(existing, spec) || h.shouldDeleteAssets(existing, spec) || h.shouldUpdateAssets(existing, spec, bucketName, config)
}

func (h *docstopicHandler) isOnPhaseChange(existing map[v1alpha1.DocsTopicSourceName]CommonAsset, status v1alpha1.CommonDocsTopicStatus) bool {
	return status.Phase != h.calculateAssetPhase(existing)
}

func (h *docstopicHandler) shouldCreateAssets(existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) bool {
	for _, source := range spec.Sources {
		if _, exists := existing[source.Name]; !exists {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) shouldUpdateAssets(existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string, config webhookconfig.AssetWebhookConfigMap) bool {
	for key, existingAsset := range existing {
		expectedSpec := findSource(spec.Sources, key, v1alpha1.DocsTopicSourceType(existingAsset.Labels[typeLabel]))
		if expectedSpec == nil {
			continue
		}

		assetWhsMap := config[expectedSpec.Type]
		expected := h.convertToCommonAssetSpec(*expectedSpec, bucketName, assetWhsMap)
		if !reflect.DeepEqual(expected, existingAsset.Spec) {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) shouldDeleteAssets(existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) bool {
	for key, existingAsset := range existing {
		if findSource(spec.Sources, key, v1alpha1.DocsTopicSourceType(existingAsset.Labels[typeLabel])) == nil {
			return true
		}
	}

	return false
}

func (h *docstopicHandler) onPhaseChange(instance ObjectMetaAccessor, status v1alpha1.CommonDocsTopicStatus, existing map[v1alpha1.DocsTopicSourceName]CommonAsset) (*v1alpha1.CommonDocsTopicStatus, error) {
	phase := h.calculateAssetPhase(existing)
	h.logInfof("Updating phase to %s", phase)

	if phase == v1alpha1.DocsTopicPending {
		h.recordNormalEventf(instance, v1alpha1.DocsTopicWaitingForAssets)
		return h.buildStatus(phase, v1alpha1.DocsTopicWaitingForAssets), nil
	}

	h.recordNormalEventf(instance, v1alpha1.DocsTopicAssetsReady)
	return h.buildStatus(phase, v1alpha1.DocsTopicAssetsReady), nil
}

func (h *docstopicHandler) onChange(ctx context.Context, instance ObjectMetaAccessor, spec v1alpha1.CommonDocsTopicSpec, status v1alpha1.CommonDocsTopicStatus, existing map[v1alpha1.DocsTopicSourceName]CommonAsset, bucketName string, cfg webhookconfig.AssetWebhookConfigMap) (*v1alpha1.CommonDocsTopicStatus, error) {
	if err := h.createMissingAssets(ctx, instance, existing, spec, bucketName, cfg); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsCreationFailed, err.Error()), status), err
	}

	if err := h.updateOutdatedAssets(ctx, instance, existing, spec, bucketName, cfg); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsUpdateFailed, err.Error()), status), err
	}

	if err := h.deleteNotExisting(ctx, instance, existing, spec); err != nil {
		return h.onFailedStatus(h.buildStatus(v1alpha1.DocsTopicFailed, v1alpha1.DocsTopicAssetsDeletionFailed, err.Error()), status), err
	}

	h.recordNormalEventf(instance, v1alpha1.DocsTopicWaitingForAssets)
	return h.buildStatus(v1alpha1.DocsTopicPending, v1alpha1.DocsTopicWaitingForAssets), nil
}

func (h *docstopicHandler) createMissingAssets(ctx context.Context, instance ObjectMetaAccessor, existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string, cfg webhookconfig.AssetWebhookConfigMap) error {
	for _, spec := range spec.Sources {
		name := spec.Name
		if _, exists := existing[name]; exists {
			continue
		}

		if err := h.createAsset(ctx, instance, spec, bucketName, cfg[spec.Type]); err != nil {
			return err
		}
	}

	return nil
}

func (h *docstopicHandler) createAsset(ctx context.Context, instance ObjectMetaAccessor, assetSpec v1alpha1.Source, bucketName string, cfg webhookconfig.AssetWebhookConfig) error {
	commonAsset := CommonAsset{
		ObjectMeta: v1.ObjectMeta{
			Name:        h.generateFullAssetName(instance.GetName(), assetSpec.Name, assetSpec.Type),
			Namespace:   instance.GetNamespace(),
			Labels:      h.buildLabels(instance.GetName(), assetSpec.Type),
			Annotations: h.buildAnnotations(assetSpec.Name),
		},
		Spec: h.convertToCommonAssetSpec(assetSpec, bucketName, cfg),
	}

	h.logInfof("Creating asset %s", commonAsset.Name)
	if err := h.assetSvc.Create(ctx, instance, commonAsset); err != nil {
		h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetCreationFailed, commonAsset.Name, err.Error())
		return err
	}
	h.logInfof("Asset %s created", commonAsset.Name)
	h.recordNormalEventf(instance, v1alpha1.DocsTopicAssetCreated, commonAsset.Name)

	return nil
}

func (h *docstopicHandler) updateOutdatedAssets(ctx context.Context, instance ObjectMetaAccessor, existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec, bucketName string, cfg webhookconfig.AssetWebhookConfigMap) error {
	for key, existingAsset := range existing {
		expectedSpec := findSource(spec.Sources, key, v1alpha1.DocsTopicSourceType(existingAsset.Labels[typeLabel]))
		if expectedSpec == nil {
			continue
		}

		h.logInfof("Updating asset %s", existingAsset.Name)
		expected := h.convertToCommonAssetSpec(*expectedSpec, bucketName, cfg[expectedSpec.Type])
		if reflect.DeepEqual(expected, existingAsset.Spec) {
			h.logInfof("Asset %s is up-to-date", existingAsset.Name)
			continue
		}

		existingAsset.Spec = expected
		if err := h.assetSvc.Update(ctx, existingAsset); err != nil {
			h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetUpdateFailed, existingAsset.Name, err.Error())
			return err
		}
		h.logInfof("Asset %s updated", existingAsset.Name)
		h.recordNormalEventf(instance, v1alpha1.DocsTopicAssetUpdated, existingAsset.Name)
	}

	return nil
}

func findSource(slice []v1alpha1.Source, sourceName v1alpha1.DocsTopicSourceName, sourceType v1alpha1.DocsTopicSourceType) *v1alpha1.Source {
	for _, source := range slice {
		if source.Name == sourceName && source.Type == sourceType {
			return &source
		}
	}
	return nil
}

func (h *docstopicHandler) deleteNotExisting(ctx context.Context, instance ObjectMetaAccessor, existing map[v1alpha1.DocsTopicSourceName]CommonAsset, spec v1alpha1.CommonDocsTopicSpec) error {
	for key, commonAsset := range existing {
		if findSource(spec.Sources, key, v1alpha1.DocsTopicSourceType(commonAsset.Labels[typeLabel])) != nil {
			continue
		}

		h.logInfof("Deleting asset %s", commonAsset.Name)
		if err := h.assetSvc.Delete(ctx, commonAsset); err != nil {
			h.recordWarningEventf(instance, v1alpha1.DocsTopicAssetDeletionFailed, commonAsset.Name, err.Error())
			return err
		}
		h.logInfof("Asset %s deleted", commonAsset.Name)
		h.recordNormalEventf(instance, v1alpha1.DocsTopicAssetDeleted, commonAsset.Name)
	}

	return nil
}

func (h *docstopicHandler) convertToAssetMap(assets []CommonAsset) map[v1alpha1.DocsTopicSourceName]CommonAsset {
	result := make(map[v1alpha1.DocsTopicSourceName]CommonAsset)

	for _, asset := range assets {
		assetShortName := asset.Annotations[assetShortNameAnnotation]
		result[v1alpha1.DocsTopicSourceName(assetShortName)] = asset
	}

	return result
}

func (h *docstopicHandler) convertToCommonAssetSpec(spec v1alpha1.Source, bucketName string, cfg webhookconfig.AssetWebhookConfig) v1alpha2.CommonAssetSpec {
	return v1alpha2.CommonAssetSpec{
		Source: v1alpha2.AssetSource{
			Mode:                     h.convertToAssetMode(spec.Mode),
			URL:                      spec.URL,
			Filter:                   spec.Filter,
			ValidationWebhookService: convertToAssetWebhookServices(cfg.Validations),
			MutationWebhookService:   convertToAssetWebhookServices(cfg.Mutations),
			MetadataWebhookService:   convertToWebhookService(cfg.MetadataExtractors),
		},
		BucketRef: v1alpha2.AssetBucketRef{
			Name: bucketName,
		},
		Parameters: spec.Parameters,
	}
}

func convertToWebhookService(services []webhookconfig.WebhookService) []v1alpha2.WebhookService {
	servicesLen := len(services)
	if servicesLen < 1 {
		return nil
	}
	result := make([]v1alpha2.WebhookService, 0, servicesLen)
	for _, service := range services {
		result = append(result, v1alpha2.WebhookService{
			Name:      service.Name,
			Namespace: service.Namespace,
			Endpoint:  service.Endpoint,
			Filter:    service.Filter,
		})
	}
	return result
}

func convertToAssetWebhookServices(services []webhookconfig.AssetWebhookService) []v1alpha2.AssetWebhookService {
	servicesLen := len(services)
	if servicesLen < 1 {
		return nil
	}
	result := make([]v1alpha2.AssetWebhookService, 0, servicesLen)
	for _, s := range services {
		result = append(result, v1alpha2.AssetWebhookService{
			WebhookService: v1alpha2.WebhookService{
				Name:      s.Name,
				Namespace: s.Namespace,
				Endpoint:  s.Endpoint,
				Filter:    s.Filter,
			},
			Parameters: s.Parameters,
		})
	}
	return result
}

func (h *docstopicHandler) buildLabels(topicName string, assetType v1alpha1.DocsTopicSourceType) map[string]string {
	labels := make(map[string]string)

	labels[docsTopicLabel] = topicName
	if assetType != "" {
		labels[typeLabel] = string(assetType)
	}

	return labels

}

func (h *docstopicHandler) buildAnnotations(assetShortName v1alpha1.DocsTopicSourceName) map[string]string {
	return map[string]string{
		assetShortNameAnnotation: string(assetShortName),
	}
}

func (h *docstopicHandler) convertToAssetMode(mode v1alpha1.DocsTopicSourceMode) v1alpha2.AssetMode {
	switch mode {
	case v1alpha1.DocsTopicIndex:
		return v1alpha2.AssetIndex
	case v1alpha1.DocsTopicPackage:
		return v1alpha2.AssetPackage
	default:
		return v1alpha2.AssetSingle
	}
}

func (h *docstopicHandler) calculateAssetPhase(existing map[v1alpha1.DocsTopicSourceName]CommonAsset) v1alpha1.DocsTopicPhase {
	for _, asset := range existing {
		if asset.Status.Phase != v1alpha2.AssetReady {
			return v1alpha1.DocsTopicPending
		}
	}

	return v1alpha1.DocsTopicReady
}

func (h *docstopicHandler) buildStatus(phase v1alpha1.DocsTopicPhase, reason v1alpha1.DocsTopicReason, args ...interface{}) *v1alpha1.CommonDocsTopicStatus {
	return &v1alpha1.CommonDocsTopicStatus{
		Phase:             phase,
		Reason:            reason,
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

func (h *docstopicHandler) recordNormalEventf(object ObjectMetaAccessor, reason v1alpha1.DocsTopicReason, args ...interface{}) {
	h.recordEventf(object, "Normal", reason, args...)
}

func (h *docstopicHandler) recordWarningEventf(object ObjectMetaAccessor, reason v1alpha1.DocsTopicReason, args ...interface{}) {
	h.recordEventf(object, "Warning", reason, args...)
}

func (h *docstopicHandler) recordEventf(object ObjectMetaAccessor, eventType string, reason v1alpha1.DocsTopicReason, args ...interface{}) {
	h.recorder.Eventf(object, eventType, reason.String(), reason.Message(), args...)
}

func (h *docstopicHandler) generateFullAssetName(docsTopicName string, assetShortName v1alpha1.DocsTopicSourceName, assetType v1alpha1.DocsTopicSourceType) string {
	assetTypeLower := strings.ToLower(string(assetType))
	assetShortNameLower := strings.ToLower(string(assetShortName))
	return h.appendSuffix(fmt.Sprintf("%s-%s-%s", docsTopicName, assetShortNameLower, assetTypeLower))
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
