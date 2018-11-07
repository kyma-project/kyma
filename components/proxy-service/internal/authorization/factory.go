package authorization

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/tokencache"
	"net/http"
)

type Strategy interface {
	Setup(r *http.Request) apperrors.AppError
	Reset()
}

type StrategyFactory interface {
	Create(credentials Credentials) Strategy
}

type authorizationStrategyFactory struct {
	oauthClient OAuthClient
}

type OAuthClient interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateTokenCache(clientID string)
}

type OauthCredentials struct {
	AuthenticationUrl string
	ClientId          string
	ClientSecret      string
}

type BasicAuthCredentials struct {
	UserName string
	Password string
}

type Credentials struct {
	Oauth *OauthCredentials
	Basic *BasicAuthCredentials
}

func (asf authorizationStrategyFactory) Create(c Credentials) Strategy {
	if c.Oauth != nil {
		oauthStrategy := newOAuthStrategy(asf.oauthClient, c.Oauth.ClientId, c.Oauth.ClientSecret, c.Oauth.AuthenticationUrl)

		return newExternalTokenStrategy(oauthStrategy)
	} else if c.Basic != nil {
		basicAuthStrategy := newBasicAuthStrategy(c.Basic.UserName, c.Basic.Password)

		return newExternalTokenStrategy(basicAuthStrategy)
	} else {
		noAuthStrategy := newNoAuthStrategy()

		return newExternalTokenStrategy(noAuthStrategy)
	}
}

type Configuration struct {
	OAuthClientTimeout int
}

func NewStrategyFactory(config Configuration) StrategyFactory {
	cache := tokencache.NewTokenCache()
	oauthClient := oauth.NewOauthClient(config.OAuthClientTimeout, cache)

	return authorizationStrategyFactory{oauthClient: oauthClient}
}
