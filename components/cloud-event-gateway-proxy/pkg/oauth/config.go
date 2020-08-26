package oauth

import (
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"golang.org/x/oauth2/clientcredentials"
)

func Config(env gateway.EnvConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     env.ClientID,
		ClientSecret: env.ClientSecret,
		TokenURL:     env.TokenEndpoint,
	}
}
