package clusterdocstopic

import (
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	ClusterDocsTopicRelistInterval time.Duration `envconfig:"default=5m"`
	ClusterBucketRegion            string        `envconfig:"optional"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
