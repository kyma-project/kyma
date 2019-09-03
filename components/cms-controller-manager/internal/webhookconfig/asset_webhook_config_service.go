package webhookconfig

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Config struct {
	CfgMapName      string `envconfig:"default=webhook-configmap"`
	CfgMapNamespace string `envconfig:"default=kyma-system"`
}

type AssetWebhookConfigMap = map[v1alpha1.DocsTopicSourceType]AssetWebhookConfig

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
	Parameters     *runtime.RawExtension `json:"parameters,omitempty"`
}

type AssetWebhookConfig struct {
	Validations        []AssetWebhookService `json:"validations,omitempty"`
	Mutations          []AssetWebhookService `json:"mutations,omitempty"`
	MetadataExtractors []WebhookService      `json:"metadataExtractors,omitempty"`
}

type assetWebhookConfigService struct {
	resourceGetter         ResourceGetter
	webhookCfgMapName      string
	webhookCfgMapNamespace string
}

//go:generate mockery -name=AssetWebhookConfigService -output=automock -outpkg=automock -case=underscore
type AssetWebhookConfigService interface {
	Get(ctx context.Context) (AssetWebhookConfigMap, error)
}

//go:generate mockery -name=ResourceGetter -output=automock -outpkg=automock -case=underscore
type ResourceGetter interface {
	Get(name string, options metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error)
}

func New(indexer ResourceGetter, webhookCfgMapName, webhookCfgMapNamespace string) *assetWebhookConfigService {
	return &assetWebhookConfigService{
		resourceGetter:         indexer,
		webhookCfgMapName:      webhookCfgMapName,
		webhookCfgMapNamespace: webhookCfgMapNamespace,
	}
}

func (r *assetWebhookConfigService) Get(ctx context.Context) (AssetWebhookConfigMap, error) {
	item, err := r.resourceGetter.Get(r.webhookCfgMapName, metav1.GetOptions{})

	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting webhook configuration in namespace %s", r.webhookCfgMapNamespace)
	}

	var cfgMap v1.ConfigMap
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.UnstructuredContent(), &cfgMap)
	if err != nil {
		return nil, errors.Wrap(err, "while converting from *unstructured.Unstructured to *v1.ConfigMap")
	}
	return toAssetWhsConfig(cfgMap)
}

func toAssetWhsConfig(configMap v1.ConfigMap) (AssetWebhookConfigMap, error) {
	result := AssetWebhookConfigMap{}
	for k, v := range configMap.Data {
		var assetWhMap AssetWebhookConfig
		if err := json.Unmarshal([]byte(v), &assetWhMap); err != nil {
			return nil, errors.Wrapf(err, "invalid content for source type type: %s", k)
		}
		result[v1alpha1.DocsTopicSourceType(k)] = assetWhMap
	}
	return result, nil
}
