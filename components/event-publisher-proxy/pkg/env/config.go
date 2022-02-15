package env

import (
	"net/http"
	"time"
)

// Config represents the environment config for the Event Publisher Proxy.
type Config struct {
	Port                int           `envconfig:"INGRESS_PORT" default:"8080"`
	ClientID            string        `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret        string        `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint       string        `envconfig:"TOKEN_ENDPOINT" required:"true"`
	EmsPublishURL       string        `envconfig:"EMS_PUBLISH_URL" required:"true"`
	MaxIdleConns        int           `envconfig:"MAX_IDLE_CONNS" default:"100"`
	MaxIdleConnsPerHost int           `envconfig:"MAX_IDLE_CONNS_PER_HOST" default:"2"`
	RequestTimeout      time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
}

// ConfigureTransport receives an HTTP transport and configure its max idle connection properties.
func (c *Config) ConfigureTransport(transport *http.Transport) {
	transport.MaxIdleConns = c.MaxIdleConns
	transport.MaxIdleConnsPerHost = c.MaxIdleConnsPerHost
}
