package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/davecgh/go-spew/spew"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger"
	"github.com/mcuadros/go-defaults"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

// Config provide application broker configuration
// Supported tags:
//	- json: 		github.com/ghodss/yaml
//	- envconfig: 	github.com/vrischmann/envconfig
//	- default: 		github.com/mcuadros/go-defaults
//	- valid         github.com/asaskevich/govalidator
// Example of valid tag: `valid:"alphanum,required"`
// Combining many tags: tags have to be separated by WHITESPACE: `json:"port" default:"8080" valid:"required"`
type Config struct {
	Logger                     logger.Config
	MockV2API                  bool             `valid:"required"`
	Port                       int              `default:"8080"`
	Storage                    []storage.Config `valid:"required"`
	BrokerRelistDurationWindow time.Duration    `valid:"required"`
	// UniqueSelectorLabelKey and UniqueSelectorLabelValue allows unambiguously distinguish AB's pods.
	UniqueSelectorLabelKey   string `valid:"required"`
	UniqueSelectorLabelValue string `valid:"required"`
	Namespace                string `valid:"required"`
	ServiceName              string `valid:"required"`
}

// Load method has following strategy:
// 1. Check env variable 'APP_CONFIG_FILE_NAME', if exists load configuration from specified file
// 2. Read configuration from environment variables (will override configuration from file)
// 3. Apply defaults
// 4. Validate
func Load(verbose bool) (*Config, error) {
	outCfg := Config{}

	cfgFile := os.Getenv("APP_CONFIG_FILE_NAME")
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, errors.Wrapf(err, "while opening config file [%s]", cfgFile)
		}
		fileConfig := Config{}
		if err := yaml.Unmarshal(b, &fileConfig); err != nil {
			return nil, errors.Wrap(err, "while unmarshalling config from file")
		}
		outCfg = fileConfig
		// fmt.Printf used, because logger will be created after reading configuration
		if verbose {
			spew.Dump("Config after applying values from file", outCfg)
		}
	}

	envConf := Config{}
	if err := envconfig.InitWithOptions(&envConf, envconfig.Options{Prefix: "APP", AllOptional: true, AllowUnexported: true}); err != nil {
		return nil, errors.Wrap(err, "while reading configuration from environment variables")
	}

	if err := mergo.MergeWithOverwrite(&outCfg, &envConf); err != nil {
		return nil, errors.Wrap(err, "while merging config from environment variables")
	}
	if verbose {
		spew.Dump("Config after applying values from environment variables", outCfg)
	}

	defaults.SetDefaults(&outCfg)

	if verbose {
		spew.Dump("Config after applying defaults", outCfg)
	}
	if _, err := govalidator.ValidateStruct(outCfg); err != nil {
		return nil, errors.Wrap(err, "while validating configuration object")
	}
	return &outCfg, nil
}
