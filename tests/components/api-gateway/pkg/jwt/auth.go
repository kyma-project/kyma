package jwt

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

func Authenticate(oauthClientID string, config oidcHydraConfig) (string, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: config.ClientConfig.TimeoutSeconds,
		Jar:     jar,
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	idTokenProvider := newOidcHydraTestFlow(httpClient, config)
	token, err := idTokenProvider.fetchIdToken()

	return token, err
}
