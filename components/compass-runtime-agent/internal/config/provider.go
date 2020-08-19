package config

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
)

const (
	connectorURLConfigKey = "CONNECTOR_URL"
	tokenConfigKey        = "TOKEN"
	runtimeIdConfigKey    = "RUNTIME_ID"
	tenantConfigKey       = "TENANT"
)

type ConnectionConfig struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorUrl"`
}

type RuntimeConfig struct {
	RuntimeId string `json:"runtimeId"`
	Tenant    string `json:"tenant"` // TODO: after full implementation of certs in Director it will no longer be needed
}

//go:generate mockery --name=Provider
type Provider interface {
	GetConnectionConfig() (ConnectionConfig, error)
	GetRuntimeConfig() (RuntimeConfig, error)
}

func NewConfigProvider(secretName types.NamespacedName, secretsRepo secrets.Repository) Provider {
	return &provider{
		secretName:  secretName,
		secretsRepo: secretsRepo,
	}
}

type provider struct {
	secretName  types.NamespacedName
	secretsRepo secrets.Repository
}

func (p *provider) GetConnectionConfig() (ConnectionConfig, error) {
	configSecret, err := p.secretsRepo.Get(p.secretName)
	if err != nil {
		return ConnectionConfig{}, errors.WithMessagef(err, "Failed to read Connection config from %s Secret", p.secretName.String())
	}

	return ConnectionConfig{
		Token:        string(configSecret[tokenConfigKey]),
		ConnectorURL: string(configSecret[connectorURLConfigKey]),
	}, nil
}

func (p *provider) GetRuntimeConfig() (RuntimeConfig, error) {
	configSecret, err := p.secretsRepo.Get(p.secretName)
	if err != nil {
		return RuntimeConfig{}, errors.WithMessagef(err, "Failed to read Runtime config from %s Secret", p.secretName)
	}

	return RuntimeConfig{
		RuntimeId: string(configSecret[runtimeIdConfigKey]),
		Tenant:    string(configSecret[tenantConfigKey]),
	}, nil
}
