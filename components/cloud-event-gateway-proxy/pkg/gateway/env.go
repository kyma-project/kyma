package gateway

type EnvConfig struct {
	Port          int    `envconfig:"INGRESS_PORT" default:"8080"`
	ClientID      string `envconfig:"CLIENT_ID" required:"true"`
	ClientSecret  string `envconfig:"CLIENT_SECRET" required:"true"`
	TokenEndpoint string `envconfig:"TOKEN_ENDPOINT" required:"true"`
	EmsCEURL      string `envconfig:"EMS_CE_URL" required:"true"`
}
