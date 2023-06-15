package eventmesh

import "fmt"

type HTTPStatusError struct {
	StatusCode int
}

func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("%v", e.StatusCode)
}

func (e *HTTPStatusError) Is(target error) bool {
	t, ok := target.(*HTTPStatusError) //nolint: errorlint // converted to pointer and checked.
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode
}

type OAuth2ClientCredentials struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	CertsURL     string
}

type Response struct {
	StatusCode int
	Error      error
}
