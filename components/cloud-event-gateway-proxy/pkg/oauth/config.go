package oauth

import (
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
)

func Config(env gateway.EnvConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     env.ClientID,
		ClientSecret: env.ClientSecret,
		TokenURL:     env.TokenEndpoint,
	}
}
