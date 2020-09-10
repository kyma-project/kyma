package oauth

import (
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/env"
)

func Config(cfg *env.Config) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenEndpoint,
	}
}
