package auth

import (
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"golang.org/x/oauth2/clientcredentials"
)

// Config returns a new oauth2 client credentials config instance.
func getDefaultOauth2Config(cfg env.Config) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenEndpoint,
	}
}
