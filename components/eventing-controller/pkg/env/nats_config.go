package env

import (
	"log"
	"regexp"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var (
	// invalidSubjectPrefixCharacters used to match and replace non-alphanumeric characters in the subject-prefix
	// as per JetStream spec https://docs.nats.io/running-a-nats-service/nats_admin/jetstream_admin/naming.
	invalidSubjectPrefixCharacters = regexp.MustCompile("[^a-zA-Z0-9.]")
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
	EnableJetStreamBackend bool `envconfig:"ENABLE_JETSTREAM_BACKEND" default:"false"`
	// Name of the JetStream stream where all events are stored.
	JSStreamName string `envconfig:"JS_STREAM_NAME" required:"true"`
	// Storage type of the stream, memory or file.
	JSStreamStorageType string `envconfig:"JS_STREAM_STORAGE_TYPE" default:"memory"`
	// Retention policy specifies when to delete events from the stream.
	//  interest: when all known observables have acknowledged a message, it can be removed.
	//  limits: messages are retained until any given limit is reached.
	//  configured via JSStreamMaxMessages and JSStreamMaxBytes.
	JSStreamRetentionPolicy string `envconfig:"JS_STREAM_RETENTION_POLICY" default:"interest"`
	JSStreamMaxMessages     int64  `envconfig:"JS_STREAM_MAX_MSGS" default:"-1"`
	JSStreamMaxBytes        int64  `envconfig:"JS_STREAM_MAX_BYTES" default:"-1"`

	// Prefix for the JetStream stream subjects filter.
	// It will be overridden by non-empty EventTypePrefix.
	JSStreamSubjectPrefix string `envconfig:"JS_STREAM_SUBJECT_PREFIX" required:"true"`

	// Deliver Policy determines for a consumer where in the stream it starts receiving messages
	// (more info https://docs.nats.io/nats-concepts/jetstream/consumers#deliverpolicy-optstartseq-optstarttime):
	// - all: The consumer starts receiving from the earliest available message.
	// - last: When first consuming messages, the consumer starts receiving messages with the latest message.
	// - last_per_subject: When first consuming messages, start with the latest one for each filtered subject
	//   currently in the stream.
	// - new: When first consuming messages, the consumer starts receiving messages that were created
	//   after the consumer was created.
	JSConsumerDeliverPolicy string `envconfig:"JS_CONSUMER_DELIVER_POLICY" default:"new"`
}

func GetNatsConfig(maxReconnects int, reconnectWait time.Duration) NatsConfig {
	cfg := NatsConfig{
		MaxReconnects: maxReconnects,
		ReconnectWait: reconnectWait,
	}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// set stream subjects prefix for JetStream
	if len(cfg.EventTypePrefix) > 0 {
		cfg.JSStreamSubjectPrefix = getCleanJSStreamSubjectPrefix(cfg.EventTypePrefix)
	}
	return cfg
}

func getCleanJSStreamSubjectPrefix(prefix string) string {
	return invalidSubjectPrefixCharacters.ReplaceAllString(prefix, "")
}
