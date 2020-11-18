package env

import "github.com/vrischmann/envconfig"

var (
	Config EnvConfig
)

type EnvConfig struct {
	InstNamespace      string `envconfig:"default=default"`
	InstResource       string `envconfig:"default=kyma-installation"`
	OverridesNamespace string `envconfig:"default=kyma-installer"`
}

func InitConfig() {
	err := envconfig.Init(&Config)
	if err != nil {
		panic(err)
	}
}
