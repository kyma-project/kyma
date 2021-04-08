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

	// EventTypePrefix prefix for the EventType
	// note: eventType format is <prefix>.<application>.<event>.<version>
	EventTypePrefix string `envconfig:"EVENT_TYPE_PREFIX" required:"true"`

	// HTTP Transport config for the message dispatcher
	MaxIdleConns        int           `envconfig:"MAX_IDLE_CONNS" default:"50"`
	MaxConnsPerHost     int           `envconfig:"MAX_CONNS_PER_HOST" default:"50"`
	MaxIdleConnsPerHost int           `envconfig:"MAX_IDLE_CONNS_PER_HOST" default:"50"`
	IdleConnTimeout     time.Duration `envconfig:"IDLE_CONN_TIMEOUT" default:"10s"`
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
