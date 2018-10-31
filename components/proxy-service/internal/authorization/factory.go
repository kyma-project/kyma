package authorization

import (
	"net/http/httputil"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"net/http"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/tokencache"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth"
)

type AuthorizationStrategy interface {
	Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError
	Reset()
}

type OAuthClient interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateTokenCache(clientID string)
}

type AuthorizationStrategyFactory struct {
	oauthClient OAuthClient
}

type Configuration struct {
	OAuthClientTimeout int
}

func  (asf AuthorizationStrategyFactory) NewOauthStrategy(clientId, clientSecret, url string) AuthorizationStrategy {
	return newOAuthStrategy(asf.oauthClient, clientId, clientSecret, url)
}

func  (asf AuthorizationStrategyFactory) NewBasicAuthStrategy(username, password string) AuthorizationStrategy {
	return newBasicAuthStrategy(username, password)
}

func NewAuthStrategyFactory(config Configuration) AuthorizationStrategyFactory {
	cache := tokencache.NewTokenCache()
	oauthClient := oauth.NewOauthClient(config.OAuthClientTimeout, cache)

	return AuthorizationStrategyFactory{oauthClient: oauthClient}
}


