package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

// TODO - test

type ConnectionConfig struct {
	Token        string `json:"token"`
	ConnectorURL string `json:"connectorUrl"`
}

type RuntimeConfig struct {
	RuntimeId string `json:"runtimeId"`
}

type config struct {
	ConnectionConfig ConnectionConfig `json:"connectionConfig"`
	RuntimeConfig    RuntimeConfig    `json:"runtimeConfig"`
}

//go:generate mockery -name=Provider
type Provider interface {
	GetConnectionConfig() (ConnectionConfig, error)
	GetRuntimeConfig() (RuntimeConfig, error)
}

func NewConfigProvider(configFile string) Provider {
	return &provider{
		configFile: configFile,
	}
}

type provider struct {
	configFile string
}

func (p *provider) GetConnectionConfig() (ConnectionConfig, error) {
	config, err := p.readConfigFile()
	if err != nil {
		return ConnectionConfig{}, errors.Wrap(err, "Failed to get Connection config")
	}

	return config.ConnectionConfig, nil
}

func (p *provider) GetRuntimeConfig() (RuntimeConfig, error) {
	config, err := p.readConfigFile()
	if err != nil {
		return RuntimeConfig{}, errors.Wrap(err, "Failed to get Runtime config")
	}

	return config.RuntimeConfig, nil
}

func (p *provider) readConfigFile() (config, error) {
	content, err := ioutil.ReadFile(p.configFile)
	if err != nil {
		return config{}, errors.Wrap(err, "Failed to read cfg file")
	}

	var cfg config
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return config{}, errors.Wrap(err, "Failed to unmarshal cfg")
	}

	return cfg, nil
}
