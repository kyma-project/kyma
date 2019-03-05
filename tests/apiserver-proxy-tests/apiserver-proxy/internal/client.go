package graphql

import (
	"crypto/tls"
	"net/http"
)

func Authenticate(config idProviderConfig) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}
