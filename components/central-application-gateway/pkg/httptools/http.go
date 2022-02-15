package httptools

import (
	"io"
	"net/http"
)

type HttpClientProvider func() HttpClient
type HttpRequestProvider func(method, url string, body io.Reader) (*http.Request, error)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func DefaultHttpClientProvider() HttpClient {
	return http.DefaultClient
}

func DefaultHttpRequestProvider(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}
