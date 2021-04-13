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
	Port           int           `envconfig:"default=80"`
	TimeoutRead    time.Duration `envconfig:"default=30s"`
	TimeoutWrite   time.Duration `envconfig:"default=30s"`
	TimeoutIdle    time.Duration `envconfig:"default=120s"`
	BusolaURL      string        `envconfig:"default=https://busola.main.hasselhoff.shoot.canary.k8s-hana.ondemand.com"`
	OIDCIssuerURL  string        `envconfig:"default=https://kyma.accounts.ondemand.com"`
	OIDCClientID   string        `envconfig:"default=6667a34d-2ea0-43fa-9b13-5ada316e5393"`
	OIDCScope      string        `envconfig:"default=openid"`
	OIDCUsePKCE    bool          `envconfig:"default=false"`
	StaticFilesDIR string        `envconfig:"optional"`
}

type configOverrides struct {
	BusolaURL     *string
	OIDCIssuerURL *string
	OIDCClientID  *string
	OIDCScope     *string
	OIDCUsePKCE   *bool
}

func LoadConfig() Config {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while processing environment variables"))
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
	if val, ok := os.LookupEnv("OVERRIDE_OIDC_ISSUER_URL"); ok {
		overrides.OIDCIssuerURL = ptr.String(val)
	}
	if val, ok := os.LookupEnv("OVERRIDE_OIDC_CLIENT_ID"); ok {
		overrides.OIDCClientID = ptr.String(val)
	}
	if val, ok := os.LookupEnv("OVERRIDE_OIDC_SCOPE"); ok {
		overrides.OIDCScope = ptr.String(val)
	}
	if val, ok := os.LookupEnv("OVERRIDE_OIDC_USE_PKCE"); ok {
		overrides.OIDCUsePKCE = ptr.BoolFromString(val)
	}

	return overrides
}

func applyOverrides(oldCfg Config, overrides configOverrides) Config {
	newCfg := oldCfg

	if overrides.BusolaURL != nil {
		newCfg.BusolaURL = *overrides.BusolaURL
	}
	if overrides.OIDCIssuerURL != nil {
		newCfg.OIDCIssuerURL = *overrides.OIDCIssuerURL
	}
	if overrides.OIDCClientID != nil {
		newCfg.OIDCClientID = *overrides.OIDCClientID
	}
	if overrides.OIDCScope != nil {
		newCfg.OIDCScope = *overrides.OIDCScope
	}
	if overrides.OIDCUsePKCE != nil {
		newCfg.OIDCUsePKCE = *overrides.OIDCUsePKCE
	}

	return newCfg
}
