package beb

import (
	"fmt"
)

const (
	MaxBEBSubscriptionNameLength = 50
)

type HTTPStatusError struct {
	StatusCode int
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("%v", e.StatusCode)
}

func (e *HTTPStatusError) Is(target error) bool {
	t, ok := target.(*HTTPStatusError)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode
}

type OAuth2ClientCredentials struct {
	ClientID     string
	ClientSecret string
}

type Response struct {
	StatusCode int
	Error      error
}
