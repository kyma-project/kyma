package authorization

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

func (e externalTokenStrategy) Setup(r *http.Request) apperrors.AppError {
	externalToken := r.Header.Get(httpconsts.HeaderAccessToken)

	if externalToken != "" {
		r.Header.Del(httpconsts.HeaderAccessToken)
		r.Header.Set(httpconsts.HeaderAuthorization, externalToken)

		return nil
	} else {
		return e.strategy.Setup(r)
	}
}

func (o externalTokenStrategy) Reset() {
	o.strategy.Reset()
}
