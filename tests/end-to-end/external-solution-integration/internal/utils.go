package internal

import (
	"crypto/tls"
	"net/http"
)

// NewHTTPClient returns new *http.Client with optional insecure SSL mode
func NewHTTPClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}
