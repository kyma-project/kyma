package endpoints

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
)

type Config struct {
	client        hydra.LoginConsentClient
	authenticator *Authenticator
	challenge     string
	state         string
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

func New(hydraAddr string, hydraPort string, authn *Authenticator) (*Config, error) {
	rawHydraURL := fmt.Sprintf("%s:%s/", hydraAddr, hydraPort)
	log.Info(rawHydraURL)
	hydraURL, err := url.Parse(rawHydraURL)
	if err != nil {
		log.Errorf("failed to parse Hydra url: %s", err)
		return nil, err
	}

	return &Config{
		client:        hydra.NewClient(&http.Client{}, *hydraURL, "https"),
		authenticator: authn,
	}, nil
}
