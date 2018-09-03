package common

import (
	"io"
	"net/http"
)

type RequestProvider func(method, url string, body io.Reader) (*http.Request, error)

func DefaultRequestProvider(method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, url, body)
}
