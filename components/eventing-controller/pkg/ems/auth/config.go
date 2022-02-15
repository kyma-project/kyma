package auth

import (
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

// Config returns a new oauth2 client credentials config instance.
func getDefaultOauth2Config(cfg env.Config) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenEndpoint,
	}
}
