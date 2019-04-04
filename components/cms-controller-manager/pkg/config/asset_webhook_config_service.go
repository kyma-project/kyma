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

type AssetWebHookConfigMap map[string]AssetWebHookConfig

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

type AssetWebHookConfig struct {
	Validations        []AssetWebhookService `json:"validations,omitempty"`
	Mutations          []AssetWebhookService `json:"mutations,omitempty"`
	MetadataExtractors []WebhookService      `json:"metadataExtractors,omitempty"`
}

type assetWebHookConfigService struct {
	client                 Client
	webhookCfgMapName      string
	webhookCfgMapNamespace string
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
}

//go:generate mockery -name=AssetWebHookConfigService -output=automock -outpkg=automock -case=underscore
type AssetWebHookConfigService interface {
	Get(ctx context.Context) (AssetWebHookConfigMap, error)
}

func NewAssetWebHookService(client Client, webhookCfgMapName, webhookCfgMapNamespace string) *assetWebHookConfigService {
	return &assetWebHookConfigService{
		client:                 client,
		webhookCfgMapName:      webhookCfgMapName,
		webhookCfgMapNamespace: webhookCfgMapNamespace,
	}
}

func (r *assetWebHookConfigService) Get(ctx context.Context) (AssetWebHookConfigMap, error) {
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

func toAssetWhsConfig(configMap v1.ConfigMap) (AssetWebHookConfigMap, error) {
	result := AssetWebHookConfigMap{}
	for k, v := range configMap.Data {
		var assetWhMap AssetWebHookConfig
		if err := json.Unmarshal([]byte(v), &assetWhMap); err != nil {
			return nil, errors.Wrapf(err, "invalid content for source type type: %s", k)
		}
		result[k] = assetWhMap
	}
	return result, nil
}
