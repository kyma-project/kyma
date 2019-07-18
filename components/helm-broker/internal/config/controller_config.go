package config

import (
	"fmt"

	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/pkg/errors"
)

// ControllerConfig provide helm broker configuration
// Supported tags:
//	- json: 		github.com/ghodss/yaml
//	- envconfig: 	github.com/vrischmann/envconfig
//	- default: 		github.com/mcuadros/go-defaults
//	- valid         github.com/asaskevich/govalidator
// Example of valid tag: `valid:"alphanum,required"`
// Combining many tags: tags have to be separated by WHITESPACE: `json:"port" default:"8080" valid:"required"`
type ControllerConfig struct {
	TmpDir                   string
	Logger                   logger.Config
	KubeconfigPath           string `envconfig:"optional"`
	Namespace                string
	ServiceName              string
	ClusterServiceBrokerName string
	Storage                  []storage.Config `valid:"required"`
	DevelopMode              bool
}

// LoadControllerConfig method has following strategy:
// 1. Check env variable 'APP_CONFIG_FILE_NAME', if exists load configuration from specified file
// 2. Read configuration from environment variables (will override configuration from file)
// 3. Apply defaults
// 4. Validate
func LoadControllerConfig(verbose bool) (*ControllerConfig, error) {
	storageConfig, err := loadStorageConfig(ControllerConfig{}, verbose)
	if err != nil {
		return nil, errors.Wrap(err, "while loading storage config")
	}

	cfg, err := initConfig(storageConfig, ControllerConfig{}, verbose)
	if err != nil {
		return nil, errors.Wrap(err, "while initiating config")
	}

	outConfig, ok := cfg.(*ControllerConfig)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, should be *Config", outConfig)
	}

	return outConfig, nil
}
