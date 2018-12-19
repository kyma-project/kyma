package authorization

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-proxy/internal/httpconsts"
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

func (b basicAuthStrategy) AddAuthorizationHeader(r *http.Request) apperrors.AppError {
	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Basic %s", basicAuth(b.username, b.password)))

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (b basicAuthStrategy) Invalidate() {
}
