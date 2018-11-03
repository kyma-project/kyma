package authentication

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"net/http"
)

type externalTokenStrategy struct {
	strategy Strategy
}

func newExternalTokenStrategy(strategy Strategy) Strategy {
	return externalTokenStrategy{strategy: strategy}
}

func (o externalTokenStrategy) Setup(r *http.Request) apperrors.AppError {
	if !handleKymaAuthentication(r) {
		return o.strategy.Setup(r)
	}

	return nil
}

func (o externalTokenStrategy) Reset() {
	o.strategy.Reset()
}

func handleKymaAuthentication(r *http.Request) bool {
	kymaAuthorization := r.Header.Get(httpconsts.HeaderAccessToken)
	if kymaAuthorization != "" {
		r.Header.Del(httpconsts.HeaderAccessToken)
		r.Header.Set(httpconsts.HeaderAuthorization, kymaAuthorization)
		return true
	}

	return false
}
