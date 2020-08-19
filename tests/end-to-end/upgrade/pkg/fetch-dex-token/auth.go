package fetch_dex_token

import (
	"crypto/tls"
	"net/http"
)

func Authenticate(config IdProviderConfig) (string, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	idTokenProvider := newDexIdTokenProvider(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()
	return token, err
}
