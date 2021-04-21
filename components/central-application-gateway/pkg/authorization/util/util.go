package util

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

func AddBasicAuthHeader(request *http.Request, clientId, clientSecret string) {
	basicAuthHeader := fmt.Sprintf("Basic %s", encodeBasicAuthCredentials(clientId, clientSecret))

	request.Header.Set(httpconsts.HeaderAuthorization, basicAuthHeader)
}

func encodeBasicAuthCredentials(username, password string) string {
	auth := username + ":" + password

	return base64.StdEncoding.EncodeToString([]byte(auth))
}
