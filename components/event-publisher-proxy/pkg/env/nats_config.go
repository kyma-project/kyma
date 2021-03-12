package env

import "time"

// Config represents the environment config for the Event Publisher NATS.
type NatsConfig struct {
	Port                 int           `envconfig:"INGRESS_PORT" default:"8080"`
	URL                  string        `envconfig:"NATS_URL" default:"nats.nats.svc.cluster.local"`
	RetryOnFailedConnect bool          `envconfig:"RETRY_ON_FAILED_CONNECT" default:"true"`
	MaxReconnects        int           `envconfig:"MAX_RECONNECTS" default:"10"`
	ReconnectWait        time.Duration `envconfig:"RECONNECT_WAIT" default:"5s"`
	RequestTimeout       time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`

	// Legacy Namespace is used as the event source for legacy events
	LegacyNamespace string `envconfig:"LEGACY_NAMESPACE" default:"kyma"`
	// LegacyEventTypePrefix is the prefix of each event as per the eventing specification, used for legacy events
	// It follows the eventType format: <LegacyEventTypePrefix>.<appName>.<event-name>.<version>
	LegacyEventTypePrefix string `envconfig:"LEGACY_EVENT_TYPE_PREFIX" default:"kyma"`
}

// Convert to a default BEB Config
func (c *NatsConfig) ToConfig() *Config {
	cfg := &Config{
		BEBNamespace:    c.LegacyNamespace,
		EventTypePrefix: c.LegacyEventTypePrefix,
	}
	return cfg
}
