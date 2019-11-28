package jwt

import (
	"crypto/tls"
	"net/http"
)

func Authenticate(config idProviderConfig) (string, error) {
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
