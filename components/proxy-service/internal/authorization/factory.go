package authorization

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization/oauth/tokencache"
	metadatamodel "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/model"
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
	Create(credentials *metadatamodel.Credentials) Strategy
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

// Create creates strategy for credentials provided
func (asf authorizationStrategyFactory) Create(c *metadatamodel.Credentials) Strategy {

	if c != nil && c.OAuth != nil {
		oauthStrategy := newOAuthStrategy(asf.oauthClient, c.OAuth.ClientID, c.OAuth.ClientSecret, c.OAuth.URL)

		return newExternalTokenStrategy(oauthStrategy)
	} else if c != nil && c.BasicAuth != nil {
		basicAuthStrategy := newBasicAuthStrategy(c.BasicAuth.Username, c.BasicAuth.Password)

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
