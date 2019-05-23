package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/internal/httptools"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
)

type externalTokenStrategy struct {
	strategy        Strategy
	headers         *map[string][]string
	queryParameters *map[string][]string
}

func newExternalTokenStrategy(strategy Strategy, headers, queryParameters *map[string][]string) Strategy {
	return externalTokenStrategy{strategy, headers, queryParameters}
}

func (e externalTokenStrategy) AddAuthorization(r *http.Request, setter TransportSetter) apperrors.AppError {
	httptools.SetHeaders(r.Header, e.headers)
	httptools.SetQueryParameters(r.URL, e.queryParameters)

	externalToken := r.Header.Get(httpconsts.HeaderAccessToken)
	if externalToken != "" {
		r.Header.Del(httpconsts.HeaderAccessToken)
		r.Header.Set(httpconsts.HeaderAuthorization, externalToken)

		return nil
	}

	return e.strategy.AddAuthorization(r, setter)
}

func (o externalTokenStrategy) Invalidate() {
	o.strategy.Invalidate()
}
