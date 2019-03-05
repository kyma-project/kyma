package authorization

import (
	"net/http"
	"net/http/httputil"

	"github.com/kyma-project/kyma/components/application-proxy/internal/authorization/util"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
)

type basicAuthStrategy struct {
	username string
	password string
}

func newBasicAuthStrategy(username, password string) basicAuthStrategy {
	return basicAuthStrategy{
		username: username,
		password: password,
	}
}

func (b basicAuthStrategy) AddAuthorization(r *http.Request, _ *httputil.ReverseProxy) apperrors.AppError {
	util.AddBasicAuthHeader(r, b.username, b.password)

	return nil
}

func (b basicAuthStrategy) Invalidate() {
}
