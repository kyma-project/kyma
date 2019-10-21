package oauth2

import (
	"fmt"
)

type TokenRequest struct {
	ID        string
	Secret    string
	Scope     string
	GrantType string
}

type tokenResponse struct {
	*Token
	*HydraError
}

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type HydraError struct {
	StatusCode       int    `json:"status_code"`
	ErrorMsg         string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (e *HydraError) Error() string {
	return fmt.Sprintf("status code: %d, error: %s, error description: %s",
		e.StatusCode,
		e.ErrorMsg,
		e.ErrorDescription)
}
