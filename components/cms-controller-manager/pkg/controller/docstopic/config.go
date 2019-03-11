package docstopic

import (
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	DocsTopicRelistInterval time.Duration `envconfig:"default=5m"`
	BucketRegion            string        `envconfig:"optional"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
