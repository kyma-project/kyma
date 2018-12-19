package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-proxy/internal/httpconsts"
)

type externalTokenStrategy struct {
	strategy Strategy
}

func newExternalTokenStrategy(strategy Strategy) Strategy {
	return externalTokenStrategy{strategy: strategy}
}

func (e externalTokenStrategy) AddAuthorizationHeader(r *http.Request) apperrors.AppError {
	externalToken := r.Header.Get(httpconsts.HeaderAccessToken)

	if externalToken != "" {
		r.Header.Del(httpconsts.HeaderAccessToken)
		r.Header.Set(httpconsts.HeaderAuthorization, externalToken)

		return nil
	} else {
		return e.strategy.AddAuthorizationHeader(r)
	}
}

func (o externalTokenStrategy) Invalidate() {
	o.strategy.Invalidate()
}
