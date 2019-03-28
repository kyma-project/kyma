package config

import (
	"context"
	"encoding/json"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AssetWebHookConfigMap map[string]AssetWebHookConfig

type AssetWebHookConfig struct {
	Validations []v1alpha2.AssetWebhookService `json:"validations,omitempty"`
	Mutations   []v1alpha2.AssetWebhookService `json:"mutations,omitempty"`
}

type assetWebHookConfigService struct {
	client Client
}

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
}

//go:generate mockery -name=AssetWebHookConfigService -output=automock -outpkg=automock -case=underscore
type AssetWebHookConfigService interface {
	Get(ctx context.Context, namespace, name string) (AssetWebHookConfigMap, error)
}

func NewAssetWebHookService(client Client) *assetWebHookConfigService {
	return &assetWebHookConfigService{
		client: client,
	}
}

func (r *assetWebHookConfigService) Get(ctx context.Context, namespace string, name string) (AssetWebHookConfigMap, error) {
	instance := &v1.ConfigMap{}
	err := r.client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, instance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting web hook configuration in namespace %s", namespace)
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
