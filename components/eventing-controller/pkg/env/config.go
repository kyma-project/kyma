package env

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the environment config for the Eventing Controller.
type Config struct {
	BebApiUrl     string `envconfig:"BEB_API_URL" default:"https://enterprise-messaging-pubsub.cfapps.sap.hana.ondemand.com/sap/ems/v1"`
	ClientID      string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret  string `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint string `envconfig:"TOKEN_ENDPOINT" required:"true"`

	WebhookActivationTimeout time.Duration `envconfig:"WEBHOOK_ACTIVATION_TIMEOUT" default:"60s"`
	WebhookClientID          string        `envconfig:"WEBHOOK_CLIENT_ID" required:"false"`
	WebhookClientSecret      string        `envconfig:"WEBHOOK_CLIENT_SECRET" required:"false"`
	WebhookTokenEndpoint     string        `envconfig:"WEBHOOK_TOKEN_ENDPOINT" required:"false"`
}

func GetConfig() *Config {
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
