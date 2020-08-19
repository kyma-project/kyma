package internal

import (
	"crypto/tls"
	"net/http"
	"time"
)

type HTTPClientOption func(client *http.Client)

// NewHTTPClient returns new *http.Client with optional insecure SSL mode
func NewHTTPClient(options ...HTTPClientOption) *http.Client {
	client := &http.Client{Timeout: 30 * time.Second}
	for _, option := range options {
		option(client)
	}
	return client
}

func WithSkipSSLVerification(skip bool) HTTPClientOption {
	return func(client *http.Client) {
		if tr, ok := client.Transport.(*http.Transport); ok {
			tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: skip}
			return
		}
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skip},
		}
	}
}

func WithClientCertificates(certs []tls.Certificate) HTTPClientOption {
	return func(client *http.Client) {
		if tr, ok := client.Transport.(*http.Transport); ok {
			tr.TLSClientConfig.Certificates = certs
			return
		}
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{Certificates: certs},
		}
	}

}
