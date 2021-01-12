package env

import "time"

// Config represents the environment config for the Event Publisher NATS.
type NatsConfig struct {
	Port                int           `envconfig:"INGRESS_PORT" default:"8080"`
	NatsPublishURL      string        `envconfig:"NATS_PUBLISH_URL" default:""`  // like in NATS event-controller
	MaxIdleConns        int           `envconfig:"MAX_IDLE_CONNS" default:"100"`  // ??
	MaxIdleConnsPerHost int           `envconfig:"MAX_IDLE_CONNS_PER_HOST" default:"2"`
	RequestTimeout      time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
	// Legacy Namespace is used as the event source for legacy events
	LegacyNamespace string `envconfig:"LEGACY_NAMESPACE" default:"kyma"`
	// LegacyEventTypePrefix is the prefix of each event as per the eventing specification, used for legacy events
	// It follows the eventType format: <LegacyEventTypePrefix>.<appName>.<event-name>.<version>
	LegacyEventTypePrefix string `envconfig:"LEGACY_EVENT_TYPE_PREFIX" default:"kyma"`
}

// Convert to a default BEB Config
func (c *NatsConfig) ToConfig() *Config{
	cfg := &Config {
		BEBNamespace: c.LegacyNamespace,
		EventTypePrefix: c.LegacyEventTypePrefix,
	}
	return cfg
}