package config

import (
	v1 "k8s.io/api/core/v1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	connectorURLConfigKey = "CONNECTOR_URL"
	tokenConfigKey        = "TOKEN"
	runtimeIdConfigKey    = "RUNTIME_ID"
	tenantConfigKey       = "TENANT"
)

//go:generate mockery -name=ConfigMapManager
type ConfigMapManager interface {
	Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error)
}

type ConnectionConfig struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorUrl"`
}

type RuntimeConfig struct {
	RuntimeId string `json:"runtimeId"`
	Tenant    string `json:"tenant"` // TODO: after full implementation of certs in Director it will no longer be needed
}

//go:generate mockery -name=Provider
type Provider interface {
	GetConnectionConfig() (ConnectionConfig, error)
	GetRuntimeConfig() (RuntimeConfig, error)
}

func NewConfigProvider(configMapName string, cmManager ConfigMapManager) Provider {
	return &provider{
		configMapName: configMapName,
		cmManager:     cmManager,
	}
}

type provider struct {
	configMapName string
	cmManager     ConfigMapManager
}

func (p *provider) GetConnectionConfig() (ConnectionConfig, error) {
	configMap, err := p.cmManager.Get(p.configMapName, metav1.GetOptions{})
	if err != nil {
		return ConnectionConfig{}, errors.WithMessagef(err, "Failed to read Connection config from %s config map", p.configMapName)
	}

	return ConnectionConfig{
		Token:        configMap.Data[tokenConfigKey],
		ConnectorURL: configMap.Data[connectorURLConfigKey],
	}, nil
}

func (p *provider) GetRuntimeConfig() (RuntimeConfig, error) {
	configMap, err := p.cmManager.Get(p.configMapName, metav1.GetOptions{})
	if err != nil {
		return RuntimeConfig{}, errors.WithMessagef(err, "Failed to read Runtime config from %s config map", p.configMapName)
	}

	return RuntimeConfig{
		RuntimeId: configMap.Data[runtimeIdConfigKey],
		Tenant:    configMap.Data[tenantConfigKey],
	}, nil
}
