package clusterdocstopic

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/webhookconfig"
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	webhookconfig.Config
	ClusterDocsTopicRelistInterval time.Duration `envconfig:"default=5m"`
	ClusterBucketRegion            string        `envconfig:"optional"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
