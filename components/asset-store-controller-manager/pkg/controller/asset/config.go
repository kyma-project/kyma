package asset

import (
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/loader"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store"
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	Store                        store.Config
	Loader                       loader.Config
	Webhook                      assethook.Config
	MaxAssetConcurrentReconciles int           `envconfig:"default=1"`
	AssetRelistInterval          time.Duration `envconfig:"default=30s"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
