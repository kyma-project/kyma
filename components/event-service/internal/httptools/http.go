package httptools

import (
	"io"
	"net/http"
)

// An HTTPClientProvider represents a function type that returns an HTTPClient
type HTTPClientProvider func() HTTPClient

// An HTTPRequestProvider represents a function type that returns an http.Request
type HTTPRequestProvider func(method, url string, body io.Reader) (*http.Request, error)

// An HTTPClient represents an interface type of HTTPClient that encapsulates a 'Do' function which takes an http.Request and returns an http.Response
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClientProvider returns a default HTTPClient implementation
func DefaultHTTPClientProvider() HTTPClient {
	return http.DefaultClient
}

// DefaultHTTPRequestProvider returns a default http.Request with the given method, url and body
func DefaultHTTPRequestProvider(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}
