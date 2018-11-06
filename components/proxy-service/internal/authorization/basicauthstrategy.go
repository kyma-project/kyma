package authorization

import (
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"net/http"
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

func (o basicAuthStrategy) Setup(r *http.Request) apperrors.AppError {
	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Basic %s", basicAuth(o.username, o.password)))

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (o basicAuthStrategy) Reset() {
}




