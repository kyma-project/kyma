package bucket

import (
	"github.com/vrischmann/envconfig"
	"time"
)

type Config struct {
	Endpoint               string        `envconfig:"default=minio.kyma.local"`
	AccessKey              string        `envconfig:""`
	SecretKey              string        `envconfig:""`
	UseSSL                 bool          `envconfig:"default=true"`
	SuccessRequeueInterval time.Duration `envconfig:"default=5m"`
	FailureRequeueInterval time.Duration `envconfig:"default=10s"`
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
