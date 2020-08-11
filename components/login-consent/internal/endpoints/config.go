package endpoints

import (
	"context"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type Config struct {
	hydraAddr     string
	hydraPort     string
	authenticator *Authenticator
}

type Authenticator struct {
	clientConfig oauth2.Config
	provider     *oidc.Provider
	ctx          context.Context
}

func NewAuthenticator(dexAddress string, clientID string, clientSecret string, redirectURL string, scopes []string) (*Authenticator, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, dexAddress)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	return &Authenticator{
		clientConfig: oauth2Config,
		provider:     provider,
		ctx:          ctx,
	}, nil
}

func NewConfig(hydraAddr string, hydraPort string, authn *Authenticator) *Config {
	return &Config{
		hydraAddr:     hydraAddr,
		hydraPort:     hydraPort,
		authenticator: authn,
	}
}
