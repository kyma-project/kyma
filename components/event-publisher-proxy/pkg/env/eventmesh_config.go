package env

import (
	"fmt"
	"net/http"
	"time"
)

// compile time check
var _ fmt.Stringer = &Config{}

// Config represents the environment config for the Event Publisher to BEB.
type Config struct {
	Port                int           `envconfig:"INGRESS_PORT" default:"8080"`
	ClientID            string        `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret        string        `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint       string        `envconfig:"TOKEN_ENDPOINT" required:"true"`
	EmsPublishURL       string        `envconfig:"EMS_PUBLISH_URL" required:"true"`
	MaxIdleConns        int           `envconfig:"MAX_IDLE_CONNS" default:"100"`
	MaxIdleConnsPerHost int           `envconfig:"MAX_IDLE_CONNS_PER_HOST" default:"2"`
	RequestTimeout      time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
	// Namespace is the name of the namespace in BEB which is used as the event source for legacy events
	Namespace string `envconfig:"BEB_NAMESPACE" required:"true"`
	// EventTypePrefix is the prefix of each event as per the eventing specification
	// It follows the eventType format: <eventTypePrefix>.<appName>.<event-name>.<version>
	EventTypePrefix string `envconfig:"EVENT_TYPE_PREFIX" default:""`
}

// ConfigureTransport receives an HTTP transport and configure its max idle connection properties.
func (c *Config) ConfigureTransport(transport *http.Transport) {
	transport.MaxIdleConns = c.MaxIdleConns
	transport.MaxIdleConnsPerHost = c.MaxIdleConnsPerHost
}

// String implements the fmt.Stringer interface
func (c *Config) String() string {
	return fmt.Sprintf("BEBConfig{ Port: %v; TokenEndPoint: %v; EmsPublishURL: %v; "+
		"MaxIdleConns: %v; MaxIdleConnsPerHost: %v; RequestTimeout: %v; Namespace: %v; EventTypePrefix: %v }",
		c.Port, c.TokenEndpoint, c.EmsPublishURL, c.MaxIdleConns, c.MaxIdleConnsPerHost, c.RequestTimeout, c.Namespace, c.EventTypePrefix)
}
