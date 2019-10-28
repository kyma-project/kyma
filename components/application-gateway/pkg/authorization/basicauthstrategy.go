package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/util"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
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

func (b basicAuthStrategy) AddAuthorization(r *http.Request, _ TransportSetter) apperrors.AppError {
	util.AddBasicAuthHeader(r, b.username, b.password)
	return nil
}

func (b basicAuthStrategy) Invalidate() {
}
