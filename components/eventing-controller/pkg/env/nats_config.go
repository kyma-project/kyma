package env

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// NatsConfig represents the environment config for the Eventing Controller with Nats.
type NatsConfig struct {
	// Following details are for eventing-controller to communicate to Nats
	Url           string `envconfig:"NATS_URL" default:"nats.nats.svc.cluster.local"`
	MaxReconnects int
	ReconnectWait time.Duration
}

func GetNatsConfig(maxReconnects int, reconnectWait time.Duration) NatsConfig {
	cfg := NatsConfig{
		MaxReconnects: maxReconnects,
		ReconnectWait: reconnectWait,
	}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
