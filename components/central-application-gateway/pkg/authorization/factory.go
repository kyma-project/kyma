package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/tokencache"
)

//go:generate mockery --name=Strategy
type Strategy interface {
	// Adds Authorization header to the request
	AddAuthorization(r *http.Request, setter clientcert.SetClientCertificateFunc, skipTLSVerification bool) apperrors.AppError
	// Invalidates internal state
	Invalidate()
}

//go:generate mockery --name=StrategyFactory
type StrategyFactory interface {
	// Creates strategy for credentials provided
	Create(credentials *Credentials) Strategy
}

//go:generate mockery --name=OAuthClient
type OAuthClient interface {
	// GetToken obtains OAuth token
	GetToken(clientID string, clientSecret string, authURL string, headers, queryParameters *map[string][]string, skipTLSVerification bool) (string, apperrors.AppError)
	GetTokenMTLS(clientID, authURL string, certificate, privateKey []byte, headers, queryParameters *map[string][]string, skipVerify bool) (string, apperrors.AppError)
	// InvalidateTokenCache resets internal token cache
	InvalidateTokenCache(clientID string, clientSecret string, authURL string)
}

type authorizationStrategyFactory struct {
	oauthClient OAuthClient
}

// Create creates strategy for credentials provided
func (asf authorizationStrategyFactory) Create(c *Credentials) Strategy {
	return newExternalTokenStrategy(asf.create(c))
}

func (asf authorizationStrategyFactory) create(c *Credentials) Strategy {
	if c != nil && c.OAuth != nil {
		return newOAuthStrategy(asf.oauthClient, c.OAuth.ClientID, c.OAuth.ClientSecret, c.OAuth.URL, c.OAuth.RequestParameters)
	} else if c != nil && c.OAuthWithCert != nil {
		oAuthStrategy := newOAuthWithCertStrategy(asf.oauthClient, c.OAuthWithCert.ClientID, c.OAuthWithCert.ClientSecret, c.OAuthWithCert.Certificate, c.OAuthWithCert.PrivateKey, c.OAuthWithCert.URL, c.OAuthWithCert.RequestParameters)
		return &oAuthStrategy
	} else if c != nil && c.BasicAuth != nil {
		return newBasicAuthStrategy(c.BasicAuth.Username, c.BasicAuth.Password)
	} else if c != nil && c.CertificateGen != nil {
		return newCertificateGenStrategy(c.CertificateGen.Certificate, c.CertificateGen.PrivateKey)
	} else {
		return newNoAuthStrategy()
	}
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
