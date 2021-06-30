package authorization

import (
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/util"
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

func (b basicAuthStrategy) AddAuthorization(r *http.Request, _ clientcert.SetClientCertificateFunc) apperrors.AppError {
	util.AddBasicAuthHeader(r, b.username, b.password)
	return nil
}

func (b basicAuthStrategy) Invalidate() {
}
