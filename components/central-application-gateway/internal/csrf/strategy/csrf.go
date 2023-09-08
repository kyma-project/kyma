package strategy

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
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

func (s *strategy) AddCSRFToken(apiRequest *http.Request, skipTLSVerify bool) apperrors.AppError {

	tokenResponse, err := s.csrfClient.GetTokenEndpointResponse(s.csrfTokenURL, s.authorizationStrategy, skipTLSVerify)
	if err != nil {
		zap.L().Error("failed to get CSRF token",
			zap.Error(err))
		return err
	}

	apiRequest.Header.Set(httpconsts.HeaderCSRFToken, tokenResponse.CSRFToken)

	mergeCookiesWithOverride(apiRequest, tokenResponse.Cookies)

	return nil
}

func (s *strategy) Invalidate() {
	s.csrfClient.InvalidateTokenCache(s.csrfTokenURL)
}

type noTokenStrategy struct{}

func (nts *noTokenStrategy) AddCSRFToken(apiRequest *http.Request, skipTLSVerify bool) apperrors.AppError {
	return nil
}

func (nts *noTokenStrategy) Invalidate() {
}

// Adds newCookies to the request. If the cookie is already present, overrides it.
func mergeCookiesWithOverride(request *http.Request, newCookies []*http.Cookie) {
	existingCookies := request.Cookies()

	for _, exCookie := range existingCookies {
		if !containsCookie(exCookie.Name, newCookies) {
			newCookies = append(newCookies, exCookie)
		}
	}

	request.Header.Del(httpconsts.HeaderCookie)

	for _, c := range newCookies {
		request.AddCookie(c)
	}
}

func containsCookie(name string, cookies []*http.Cookie) bool {
	for _, c := range cookies {
		if c.Name == name {
			return true
		}
	}

	return false
}
