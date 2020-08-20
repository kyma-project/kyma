package ems

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"

	"golang.org/x/oauth2"
)

type Token struct {
	AccessToken string `json:"access_token"`

	TokenType string `json:"token_type,omitempty"`

	Expiry int64 `json:"expires_in,omitempty"`

	JTI string `json:"jti,omitempty"`

	Scope string `json:"scope,omitempty"`
}

func (t Token) ToOAuthAccessToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
		Expiry:      time.Now().Add(time.Duration(t.Expiry) * time.Second),
		TokenType:   t.TokenType,
	}
}

func FetchOAuthToken(env gateway.EnvConfig) (*oauth2.Token, error) {
	req, err := http.NewRequest(http.MethodPost, env.TokenEndpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed in creating a New request")
	}
	req.SetBasicAuth(env.ClientID, env.ClientSecret)
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	defer client.CloseIdleConnections()

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed in client.Do")
	}
	defer func() { _ = resp.Body.Close() }()

	var emsToken = &Token{}
	if err = json.NewDecoder(resp.Body).Decode(emsToken); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}
	return emsToken.ToOAuthAccessToken(), nil
}
