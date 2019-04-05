package config

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AssetWebhookConfigMap map[string]AssetWebhookConfig

type WebhookService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// +optional
	Filter string `json:"filter,omitempty"`
}

type AssetWebhookService struct {
	WebhookService `json:",inline"`
	Metadata       *runtime.RawExtension `json:"metadata,omitempty"`
}

type AssetWebhookConfig struct {
	Validations        []AssetWebhookService `json:"validations,omitempty"`
	Mutations          []AssetWebhookService `json:"mutations,omitempty"`
	MetadataExtractors []WebhookService      `json:"metadataExtractors,omitempty"`
}

type assetWebhookConfigService struct {
	client                 Client
	webhookCfgMapName      string
	webhookCfgMapNamespace string
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
}

//go:generate mockery -name=AssetWebhookConfigService -output=automock -outpkg=automock -case=underscore
type AssetWebhookConfigService interface {
	Get(ctx context.Context) (AssetWebhookConfigMap, error)
}

func NewAssetWebhookService(client Client, webhookCfgMapName, webhookCfgMapNamespace string) *assetWebhookConfigService {
	return &assetWebhookConfigService{
		client:                 client,
		webhookCfgMapName:      webhookCfgMapName,
		webhookCfgMapNamespace: webhookCfgMapNamespace,
	}
}

func (r *assetWebhookConfigService) Get(ctx context.Context) (AssetWebhookConfigMap, error) {
	instance := &v1.ConfigMap{}
	err := r.client.Get(ctx, types.NamespacedName{Name: r.webhookCfgMapName, Namespace: r.webhookCfgMapNamespace}, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting web hook configuration in namespace %s", r.webhookCfgMapNamespace)
	}
	return toAssetWhsConfig(*instance)
}

func toAssetWhsConfig(configMap v1.ConfigMap) (AssetWebhookConfigMap, error) {
	result := AssetWebhookConfigMap{}
	for k, v := range configMap.Data {
		var assetWhMap AssetWebhookConfig
		if err := json.Unmarshal([]byte(v), &assetWhMap); err != nil {
			return nil, errors.Wrapf(err, "invalid content for source type type: %s", k)
		}
		result[k] = assetWhMap
	}
	return result, nil
}
