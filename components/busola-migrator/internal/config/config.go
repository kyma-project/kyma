package config

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Port         int           `envconfig:"default=80"`
	TimeoutRead  time.Duration `envconfig:"default=30s"`
	TimeoutWrite time.Duration `envconfig:"default=30s"`
	TimeoutIdle  time.Duration `envconfig:"default=120s"`
	BusolaURL    string        `envconfig:"default=https://google.com/"`
}

func LoadConfig() Config {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to process environment variables"))
	}

	return cfg
}
