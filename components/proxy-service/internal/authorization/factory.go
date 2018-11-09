package authorization

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/tokencache"
	"net/http"
)

type Strategy interface {
	// Adds Authorization header to the request
	AddAuthorizationHeader(r *http.Request) apperrors.AppError
	// Invalidates internal state
	Invalidate()
}

type StrategyFactory interface {
	// Creates strategy for credentials provided
	Create(credentials Credentials) Strategy
}

type OAuthClient interface {
	// GetToken obtains OAuth token
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	// InvalidateTokenCache resets internal token cache
	InvalidateTokenCache(clientID string)
}

type authorizationStrategyFactory struct {
	oauthClient OAuthClient
}

type OauthCredentials struct {
	Url          string
	ClientId     string
	ClientSecret string
}

type BasicAuthCredentials struct {
	Username string
	Password string
}

type Credentials struct {
	Oauth *OauthCredentials
	Basic *BasicAuthCredentials
}

// Create creates strategy for credentials provided
func (asf authorizationStrategyFactory) Create(c Credentials) Strategy {
	if c.Oauth != nil {
		oauthStrategy := newOAuthStrategy(asf.oauthClient, c.Oauth.ClientId, c.Oauth.ClientSecret, c.Oauth.Url)

		return newExternalTokenStrategy(oauthStrategy)
	} else if c.Basic != nil {
		basicAuthStrategy := newBasicAuthStrategy(c.Basic.Username, c.Basic.Password)

		return newExternalTokenStrategy(basicAuthStrategy)
	} else {
		noAuthStrategy := newNoAuthStrategy()

		return newExternalTokenStrategy(noAuthStrategy)
	}
}

// Factory configuration options
type FactoryConfiguration struct {
	OAuthClientTimeout int
}

// NewStrategyFactory creates factory for instantiating Strategy implementations
func NewStrategyFactory(config FactoryConfiguration) StrategyFactory {
	cache := tokencache.NewTokenCache()
	oauthClient := oauth.NewOauthClient(config.OAuthClientTimeout, cache)

	return authorizationStrategyFactory{oauthClient: oauthClient}
}
