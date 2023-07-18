package env

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// NATSConfig represents the environment config for the Eventing Controller with Nats.
type NATSConfig struct {
	// Following details are for eventing-controller to communicate to Nats
	URL           string `envconfig:"NATS_URL"`
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
	// Name of the JetStream stream where all events are stored.
	JSStreamName string `envconfig:"JS_STREAM_NAME" required:"true"`
	// Prefix for the subjects in the stream.
	JSSubjectPrefix string `envconfig:"JS_STREAM_SUBJECT_PREFIX" required:"true"`
	// Storage type of the stream, memory or file.
	JSStreamStorageType string `envconfig:"JS_STREAM_STORAGE_TYPE" default:"memory"`
	// Number of replicas for the JetStream stream
	JSStreamReplicas int `envconfig:"JS_STREAM_REPLICAS" default:"1"`
	// Retention policy specifies when to delete events from the stream.
	//  interest: when all known observables have acknowledged a message, it can be removed.
	//  limits: messages are retained until any given limit is reached.
	//  configured via JSStreamMaxMessages and JSStreamMaxBytes.
	JSStreamRetentionPolicy string `envconfig:"JS_STREAM_RETENTION_POLICY" default:"interest"`
	JSStreamMaxMessages     int64  `envconfig:"JS_STREAM_MAX_MSGS" default:"-1"`
	JSStreamMaxBytes        string `envconfig:"JS_STREAM_MAX_BYTES" default:"-1"`
	JSStreamMaxMsgsPerTopic int64  `envconfig:"JS_STREAM_MAX_MSGS_PER_TOPIC" default:"-1"`
	// JSStreamDiscardPolicy specifies which events to discard from the stream in case limits are reached
	//  new: reject new messages for the stream
	//  old: discard old messages from the stream to make room for new messages
	JSStreamDiscardPolicy string `envconfig:"JS_STREAM_DISCARD_POLICY" default:"new"`
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

func GetNATSConfig(maxReconnects int, reconnectWait time.Duration) (NATSConfig, error) {
	cfg := NATSConfig{
		MaxReconnects: maxReconnects,
		ReconnectWait: reconnectWait,
	}
	if err := envconfig.Process("", &cfg); err != nil {
		return NATSConfig{}, err
	}
	return cfg, nil
}
