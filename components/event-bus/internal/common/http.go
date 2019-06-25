package common

import (
	"io"
	"net/http"
)

// RequestProvider represents an HTTP request provider.
type RequestProvider func(method, url string, body io.Reader) (*http.Request, error)
