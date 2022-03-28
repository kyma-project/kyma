package env

import (
	"log"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// NatsConfig represents the environment config for the Eventing Controller with Nats.
type NatsConfig struct {
	// Following details are for eventing-controller to communicate to Nats
	URL           string `envconfig:"NATS_URL" default:"nats.nats.svc.cluster.local"`
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

	// JetStream-specific configs
	// Storage type of the stream, memory or file.
	JSStreamStorageType string `envconfig:"JS_STREAM_STORAGE_TYPE" default:"memory"`
	// Retention policy specifies when to delete events from the stream.
	//  interest: when all known observables have acknowledged a message, it can be removed.
	//  limits: messages are retained until any given limit is reached.
	//  configured via JSStreamMaxMessages and JSStreamMaxBytes.
	JSStreamRetentionPolicy string `envconfig:"JS_STREAM_RETENTION_POLICY" default:"interest"`
	JSStreamMaxMessages     int64  `envconfig:"JS_STREAM_MAX_MSGS" default:"-1"`
	JSStreamMaxBytes        int64  `envconfig:"JS_STREAM_MAX_BYTES" default:"-1"`

	// Name of the JetStream stream where all events are stored.
	JSStreamName string `default:"kyma"`
}

func GetNatsConfig(maxReconnects int, reconnectWait time.Duration) NatsConfig {
	cfg := NatsConfig{
		MaxReconnects: maxReconnects,
		ReconnectWait: reconnectWait,
	}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// set stream name for JetStream
	if len(cfg.EventTypePrefix) > 0 {
		cfg.JSStreamName = getStreamNameForJetStream(cfg.EventTypePrefix)
	}
	return cfg
}

func getStreamNameForJetStream(eventTypePrefix string) string {
	return strings.ToLower(strings.Split(eventTypePrefix, ".")[0])
}
