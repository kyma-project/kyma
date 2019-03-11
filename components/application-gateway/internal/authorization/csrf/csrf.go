package csrf

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization"
	log "github.com/sirupsen/logrus"
)

type TokenStrategyFactory interface {
	Create(authorizationStrategy authorization.Strategy, csrfTokenEndpointURL string) TokenStrategy
}

type TokenStrategy interface {
	//Sets CSRF Token into requests to external APIs
	AddCSRFToken(apiRequest *http.Request) apperrors.AppError

	//Invalidates cached CSRF Token
	Invalidate()
}

type strategyFactory struct {
	csrfClient *Client
}

type strategy struct {
	authorizationStrategy authorization.Strategy
	csrfTokenURL          string
	csrfClient            *Client
}

func NewTokenStrategyFactory(csrfClient *Client) TokenStrategyFactory {
	return &strategyFactory{csrfClient}
}

func (tsf *strategyFactory) Create(authorizationStrategy authorization.Strategy, csrfTokenEndpointURL string) TokenStrategy {
	if csrfTokenEndpointURL == "" {
		//TODO: DEBUG
		log.Print("CSRF Token not configured!")
		return &noTokenStrategy{}
	}
	return &strategy{authorizationStrategy, csrfTokenEndpointURL, tsf.csrfClient}
}

func (s *strategy) AddCSRFToken(apiRequest *http.Request) apperrors.AppError {

	resp, err := s.csrfClient.GetTokenEndpointResponse(s.csrfTokenURL, s.authorizationStrategy)
	if err != nil {
		log.Errorf("failed to get CSRF token : '%s'", err)
		return err
	}

	apiRequest.Header.Set(httpconsts.HeaderCSRFToken, resp.csrfToken)
	apiRequest.Header.Set(httpconsts.HeaderCookie, resp.cookie)

	return nil
}

func (s *strategy) Invalidate() {
	s.csrfClient.invalidateTokenCache(s.csrfTokenURL)
}

type noTokenStrategy struct{}

func (nts *noTokenStrategy) AddCSRFToken(apiRequest *http.Request) apperrors.AppError {
	return nil
}

func (nts *noTokenStrategy) Invalidate() {
}
