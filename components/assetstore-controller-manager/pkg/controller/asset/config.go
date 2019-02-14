package asset

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	Store                store.Config
	AssetRequeueInterval time.Duration `envconfig:"default=5m"`
	TemporaryDirectory   string        `envconfig:"default=/tmp"`
	MutationTimeout      time.Duration `envconfig:"default=1m"`
	ValidationTimeout    time.Duration `envconfig:"default=1m"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
