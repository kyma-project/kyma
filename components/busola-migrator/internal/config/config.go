package config

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Port           int           `envconfig:"default=80"`
	TimeoutRead    time.Duration `envconfig:"default=30s"`
	TimeoutWrite   time.Duration `envconfig:"default=30s"`
	TimeoutIdle    time.Duration `envconfig:"default=120s"`
	BusolaURL      string        `envconfig:"default=https://busola.main.hasselhoff.shoot.canary.k8s-hana.ondemand.com"`
	OIDCIssuerURL  string        `envconfig:"default=https://kyma.accounts.ondemand.com/"`
	OIDCClientID   string        `envconfig:"default=6667a34d-2ea0-43fa-9b13-5ada316e5393"`
	OIDCScope      string        `envconfig:"default=openid"`
	OIDCUsePKCE    bool          `envconfig:"default=false"`
	StaticFilesDIR string        `envconfig:"optional"`
}

func LoadConfig() Config {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while processing environment variables"))
	}

	return cfg
}
