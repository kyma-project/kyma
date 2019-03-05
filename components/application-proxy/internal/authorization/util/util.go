package util

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-proxy/internal/httpconsts"
)

func AddBasicAuthHeader(request *http.Request, clientId, clientSecret string) {
	basicAuthHeader := fmt.Sprintf("Basic %s", basicAuth(clientId, clientSecret))

	request.Header.Set(httpconsts.HeaderAuthorization, basicAuthHeader)
}

func basicAuth(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}
