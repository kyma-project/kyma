package authorization

import (
	"net/http"
	"net/http/httputil"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
)

func newNoAuthStrategy() noAuthStrategy {
	return noAuthStrategy{}
}

type noAuthStrategy struct {
}

func (ns noAuthStrategy) AddAuthorization(_ *http.Request, _ *httputil.ReverseProxy) apperrors.AppError {
	return nil
}

func (ns noAuthStrategy) Invalidate() {

}
