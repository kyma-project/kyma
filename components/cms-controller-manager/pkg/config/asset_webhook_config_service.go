package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	indexer                Indexer
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

//go:generate mockery -name=Indexer -output=automock -outpkg=automock -case=underscore
type Indexer interface {
	GetByKey(key string) (item interface{}, exists bool, err error)
}

func New(indexer Indexer, webhookCfgMapName, webhookCfgMapNamespace string) *assetWebhookConfigService {
	return &assetWebhookConfigService{
		indexer:                indexer,
		webhookCfgMapName:      webhookCfgMapName,
		webhookCfgMapNamespace: webhookCfgMapNamespace,
	}
}

func (r *assetWebhookConfigService) Get(ctx context.Context) (AssetWebhookConfigMap, error) {
	key := fmt.Sprintf("%s/%s", r.webhookCfgMapNamespace, r.webhookCfgMapName)
	item, exists, err := r.indexer.GetByKey(key)
	if err != nil || !exists {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting webhook configuration in namespace %s", r.webhookCfgMapNamespace)
	}
	cfgMap, ok := item.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1.ConfigMap", item)
	}
	return toAssetWhsConfig(*cfgMap)
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
