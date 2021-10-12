package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type externalTokenStrategy struct {
	strategy Strategy
}

func newExternalTokenStrategy(strategy Strategy) Strategy {
	return externalTokenStrategy{strategy}
}

func (e externalTokenStrategy) AddAuthorization(r *http.Request, setter clientcert.SetClientCertificateFunc) apperrors.AppError {
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
