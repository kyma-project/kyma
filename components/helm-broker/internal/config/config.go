package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kyma-project/kyma/components/helm-broker/internal/helm"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	defaults "github.com/mcuadros/go-defaults"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config provide helm broker configuration
// Supported tags:
//	- json: 		github.com/ghodss/yaml
//	- envconfig: 	github.com/vrischmann/envconfig
//	- default: 		github.com/mcuadros/go-defaults
//	- valid         github.com/asaskevich/govalidator
// Example of valid tag: `valid:"alphanum,required"`
// Combining many tags: tags have to be separated by WHITESPACE: `json:"port" default:"8080" valid:"required"`
type Config struct {
	Logger         logger.Config
	KubeconfigPath string `envconfig:"optional"`
	// TmpDir defines temporary directory path where bundles .tgz files will be extracted
	TmpDir                   string
	Namespace                string
	Port                     int              `default:"8080"`
	Storage                  []storage.Config `valid:"required"`
	Helm                     helm.Config      `valid:"required"`
	ClusterServiceBrokerName string
	HelmBrokerURL            string
	DevelopMode              bool
}

// Load method has following strategy:
// 1. Check env variable 'APP_CONFIG_FILE_NAME', if exists load configuration from specified file
// 2. Read configuration from environment variables (will override configuration from file)
// 3. Apply defaults
// 4. Validate
func Load(verbose bool) (*Config, error) {
	storageConfig, err := loadStorageConfig(Config{}, verbose)
	if err != nil {
		return nil, errors.Wrap(err, "while loading storage config")
	}

	cfg, err := initConfig(storageConfig, Config{}, verbose)
	if err != nil {
		return nil, errors.Wrap(err, "while initiating config")
	}

	outConfig, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, should be *Config", outConfig)
	}

	return outConfig, nil
}

func initConfig(storageConfig interface{}, config interface{}, verbose bool) (interface{}, error) {
	if err := envconfig.InitWithOptions(&config, envconfig.Options{Prefix: "APP", AllOptional: true, AllowUnexported: true}); err != nil {
		return nil, errors.Wrap(err, "while reading configuration from environment variables")
	}

	if err := mergo.MergeWithOverwrite(&storageConfig, &config); err != nil {
		return nil, errors.Wrap(err, "while merging config from environment variables")
	}
	if verbose {
		fmt.Printf("Config after applying values from environment variables: %+v\n", storageConfig)
	}

	defaults.SetDefaults(&storageConfig)

	if verbose {
		fmt.Printf("Config after applying defaults: %+v\n", config)
	}
	if _, err := govalidator.ValidateStruct(config); err != nil {
		return nil, errors.Wrap(err, "while validating configuration object")
	}
	return config, nil
}

func loadStorageConfig(cfg interface{}, verbose bool) (interface{}, error) {
	cfgFile := os.Getenv("APP_CONFIG_FILE_NAME")
	fileConfig := Config{}
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, errors.Wrapf(err, "while opening config file [%s]", cfgFile)
		}
		if err := yaml.Unmarshal(b, &fileConfig); err != nil {
			return nil, errors.Wrap(err, "while unmarshalling config from file")
		}
		if verbose {
			fmt.Printf("Config after applying values from file: %+v\n", cfg)
		}
	}
	return fileConfig, nil
}
