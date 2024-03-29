package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
)

func newNoAuthStrategy() noAuthStrategy {
	return noAuthStrategy{}
}

type noAuthStrategy struct {
}

func (ns noAuthStrategy) AddAuthorization(_ *http.Request, _ clientcert.SetClientCertificateFunc, _ bool) apperrors.AppError {
	return nil
}

func (ns noAuthStrategy) Invalidate() {

}
