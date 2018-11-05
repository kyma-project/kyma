package authentication

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authentication/oauth"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authentication/oauth/tokencache"
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

func (asf authorizationStrategyFactory) Create(credentials Credentials) Strategy {
	if credentials.Oauth != nil {
		clientId := credentials.Oauth.ClientId
		clientSecret := credentials.Oauth.ClientSecret
		url := credentials.Oauth.AuthenticationUrl

		oauthStrategy := newOAuthStrategy(asf.oauthClient, clientId, clientSecret, url)

		return newExternalTokenStrategy(oauthStrategy)
	} else if credentials.Basic != nil {
		username := credentials.Basic.UserName
		password := credentials.Basic.Password

		basicAuthStrategy := newBasicAuthStrategy(username, password)

		return newExternalTokenStrategy(basicAuthStrategy)
	} else {
		return newExternalTokenStrategy(noneStrategy{})
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

type noneStrategy struct {
}

func (ns noneStrategy) Setup(r *http.Request) apperrors.AppError {
	return nil
}

func (ns noneStrategy) Reset() {

}
