package config

import (
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/components/busola-migrator/pkg/ptr"
)

type Config struct {
	Domain         string        `envconfig:"default=localhost"`
	Port           int           `envconfig:"default=80"`
	TimeoutRead    time.Duration `envconfig:"default=30s"`
	TimeoutWrite   time.Duration `envconfig:"default=30s"`
	TimeoutIdle    time.Duration `envconfig:"default=120s"`
	BusolaURL      string        `envconfig:"default=https://busola.main.hasselhoff.shoot.canary.k8s-hana.ondemand.com"`
	StaticFilesDIR string        `envconfig:"optional"`
	KubeconfigID   string        `envconfig:""`

	UAA UAAConfig
}

type UAAConfig struct {
	Enabled      bool   `envconfig:"default=true"`
	URL          string `envconfig:"optional"`
	ClientID     string `envconfig:"optional"`
	ClientSecret string `envconfig:"optional"`
	RedirectURI  string `envconfig:"-"`
}

type configOverrides struct {
	BusolaURL *string
}

func LoadConfig() Config {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while processing environment variables"))
	}

	if cfg.UAA.URL == "" || cfg.UAA.ClientID == "" || cfg.UAA.ClientSecret == "" {
		cfg.UAA.Enabled = false
	}
	if !cfg.UAA.Enabled {
		log.Println("UAA Migrator functionality disabled.")
	}

	overrides := getOverrides()

	cfg = applyOverrides(cfg, overrides)

	return cfg
}

func getOverrides() configOverrides {
	var overrides configOverrides

	if val, ok := os.LookupEnv("OVERRIDE_BUSOLA_URL"); ok {
		overrides.BusolaURL = ptr.String(val)
	}
	return overrides
}

func applyOverrides(oldCfg Config, overrides configOverrides) Config {
	newCfg := oldCfg
	if overrides.BusolaURL != nil {
		newCfg.BusolaURL = *overrides.BusolaURL
	}
	return newCfg
}
