package authentication

import (
	"crypto/tls"
	"net/http"
)

func Authenticate(config IdProviderConfig) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: config.ClientConfig.TimeoutSeconds,
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}
