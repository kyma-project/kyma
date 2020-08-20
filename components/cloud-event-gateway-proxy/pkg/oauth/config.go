package oauth

import (
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"golang.org/x/oauth2"
)

func Config(env gateway.EnvConfig) oauth2.Config {
	return oauth2.Config{
		ClientID:     env.ClientID,
		ClientSecret: env.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: env.TokenEndpoint,
		},
	}
}
