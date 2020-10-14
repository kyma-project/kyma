package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Authenticator struct {
	config *Config
}

type AccessToken struct {
	Value string `json:"access_token"`
}

func NewAuthenticator(config *Config) *Authenticator {
	return &Authenticator{config: config}
}

func (a *Authenticator) Authenticate() (*AccessToken, error) {
	req, err := http.NewRequest(http.MethodPost, a.config.TokenEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Accept", "application/json")

	client := http.Client{}
	defer client.CloseIdleConnections()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp == nil {
		return nil, fmt.Errorf("could not unmarshal response: %v", resp)
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to authenticate with error: %v; %v", resp.StatusCode, resp.Status)
	}

	token := new(AccessToken)
	if err = json.NewDecoder(resp.Body).Decode(token); err != nil {
		return nil, err
	}

	return token, nil
}
