package strategy

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	log "github.com/sirupsen/logrus"
)

func NewTokenStrategyFactory(csrfClient csrf.Client) csrf.TokenStrategyFactory {
	return &strategyFactory{csrfClient}
}

type strategyFactory struct {
	csrfClient csrf.Client
}

func (tsf *strategyFactory) Create(authorizationStrategy authorization.Strategy, csrfTokenEndpointURL string) csrf.TokenStrategy {
	if csrfTokenEndpointURL == "" {
		return &noTokenStrategy{}
	}
	return &strategy{authorizationStrategy, csrfTokenEndpointURL, tsf.csrfClient}
}

type strategy struct {
	authorizationStrategy authorization.Strategy
	csrfTokenURL          string
	csrfClient            csrf.Client
}

func (s *strategy) AddCSRFToken(apiRequest *http.Request) apperrors.AppError {

	tokenResponse, err := s.csrfClient.GetTokenEndpointResponse(s.csrfTokenURL, s.authorizationStrategy)
	if err != nil {
		log.Errorf("failed to get CSRF token : '%s'", err)
		return err
	}

	apiRequest.Header.Set(httpconsts.HeaderCSRFToken, tokenResponse.CSRFToken)
	for _, cookie := range tokenResponse.Cookies {
		apiRequest.AddCookie(cookie)
	}

	return nil
}

func (s *strategy) Invalidate() {
	s.csrfClient.InvalidateTokenCache(s.csrfTokenURL)
}

type noTokenStrategy struct{}

func (nts *noTokenStrategy) AddCSRFToken(apiRequest *http.Request) apperrors.AppError {
	return nil
}

func (nts *noTokenStrategy) Invalidate() {
}
