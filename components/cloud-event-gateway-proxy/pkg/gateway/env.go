package gateway

import "net/http"

type EnvConfig struct {
	Port                int    `envconfig:"INGRESS_PORT" default:"8080"`
	ClientID            string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret        string `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint       string `envconfig:"TOKEN_ENDPOINT" required:"true"`
	EmsCEURL            string `envconfig:"EMS_CE_URL" required:"true"`
	MaxIdleConns        int    `envconfig:"MAX_IDLE_CONNS" default:"100"`
	MaxIdleConnsPerHost int    `envconfig:"MAX_IDLE_CONNS_PER_HOST" default:"2"`
}

func (e *EnvConfig) ConfigureTransport(transport *http.Transport) {
	transport.MaxIdleConns = e.MaxIdleConns
	transport.MaxIdleConnsPerHost = e.MaxIdleConnsPerHost
}
