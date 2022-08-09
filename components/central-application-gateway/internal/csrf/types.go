package csrf

import (
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
)

//CSRF Client is an HTTP client responsible for fetching and caching CSRF Tokens.
//go:generate mockery --name=Client
type Client interface {
	//Fetches data from CSRF Token Endpoint
	GetTokenEndpointResponse(csrfEndpointURL string, strategy authorization.Strategy, skipTLSVerify bool) (*Response, apperrors.AppError)

	//Invalidates cached data
	InvalidateTokenCache(csrfEndpointURL string)
}

//CSFR Endpoint response data
type Response struct {
	CSRFToken string         //Opaque value
	Cookies   []*http.Cookie //Must be included in API requests along with the token for CSFR verification to succeed
}

//Creates new instances of TokenStrategy
//go:generate mockery --name=TokenStrategyFactory
type TokenStrategyFactory interface {
	Create(authorizationStrategy authorization.Strategy, csrfTokenEndpointURL string) TokenStrategy
}

//Augments upstream API requests with CSRF data.
//go:generate mockery --name=TokenStrategy
type TokenStrategy interface {
	//Sets CSRF Token into requests to external APIs
	AddCSRFToken(apiRequest *http.Request, skipTLSVerify bool) apperrors.AppError

	//Invalidates cached CSRF Token
	Invalidate()
}
