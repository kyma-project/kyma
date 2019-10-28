package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/oauth"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/oauth/tokencache"
)

type Strategy interface {
	// Adds Authorization header to the request
	AddAuthorization(r *http.Request, setter TransportSetter) apperrors.AppError
	// Invalidates internal state
	Invalidate()
}

type TransportSetter func(transport *http.Transport)

type StrategyFactory interface {
	// Creates strategy for credentials provided
	Create(credentials *Credentials) Strategy
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
func (asf authorizationStrategyFactory) Create(c *Credentials) Strategy {
	var strategy Strategy

	if c != nil && c.OAuth != nil {
		strategy = newOAuthStrategy(asf.oauthClient, c.OAuth.ClientID, c.OAuth.ClientSecret, c.OAuth.URL)
	} else if c != nil && c.BasicAuth != nil {
		strategy = newBasicAuthStrategy(c.BasicAuth.Username, c.BasicAuth.Password)
	} else if c != nil && c.CertificateGen != nil {
		strategy = newCertificateGenStrategy(c.CertificateGen.Certificate, c.CertificateGen.PrivateKey)
	} else {
		strategy = newNoAuthStrategy()
	}

	return newExternalTokenStrategy(strategy)
}

// FactoryConfiguration holds factory configuration options
type FactoryConfiguration struct {
	OAuthClientTimeout int
}

// NewStrategyFactory creates factory for instantiating Strategy implementations
func NewStrategyFactory(config FactoryConfiguration) StrategyFactory {
	cache := tokencache.NewTokenCache()
	oauthClient := oauth.NewOauthClient(config.OAuthClientTimeout, cache)

	return authorizationStrategyFactory{oauthClient: oauthClient}
}
