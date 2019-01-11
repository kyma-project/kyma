package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
)

func newNoAuthStrategy() noAuthStrategy {
	return noAuthStrategy{}
}

type noAuthStrategy struct {
}

func (ns noAuthStrategy) AddAuthorizationHeader(r *http.Request) apperrors.AppError {
	return nil
}

func (ns noAuthStrategy) Invalidate() {

}
