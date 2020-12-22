package env

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config represents the environment config for the Eventing Controller with NATS.
type NatsConfig struct {
	// Following details are for eventing-controller to communicate to BEB
	Url string `envconfig:"NATS_URL" default:"nats.nats.svc.cluster.local"`
}

func GetNatsConfig() NatsConfig {
	cfg := NatsConfig{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}
	return cfg
}
