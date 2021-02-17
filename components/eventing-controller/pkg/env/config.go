package env

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the environment config for the Eventing Controller.
type Config struct {
	// Following details are for eventing-controller to communicate to BEB
	BebApiUrl     string `envconfig:"BEB_API_URL" default:"https://enterprise-messaging-pubsub.cfapps.sap.hana.ondemand.com/sap/ems/v1"`
	ClientID      string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret  string `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint string `envconfig:"TOKEN_ENDPOINT" required:"true"`

	// Following details are for BEB to communicate to Kyma
	WebhookActivationTimeout time.Duration `envconfig:"WEBHOOK_ACTIVATION_TIMEOUT" default:"60s"`
	WebhookClientID          string        `envconfig:"WEBHOOK_CLIENT_ID" required:"true"`
	WebhookClientSecret      string        `envconfig:"WEBHOOK_CLIENT_SECRET" required:"true"`
	WebhookTokenEndpoint     string        `envconfig:"WEBHOOK_TOKEN_ENDPOINT" required:"true"`

	// Domain holds the Kyma domain
	Domain string `envconfig:"DOMAIN" required:"true"`

	// EventTypePrefix prefix for the EventType
	// note: eventType format is <prefix>.<application>.<event>.<version>
	EventTypePrefix string `envconfig:"EVENT_TYPE_PREFIX" required:"true"`
}

func GetConfig() Config {
	cfg := Config{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
