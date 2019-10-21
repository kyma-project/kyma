package oauth2

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/avast/retry-go"

	"github.com/pkg/errors"
)

const (
	tokenEndpoint = "oauth2/token"
)

type tokenManager struct {
	httpClient *http.Client
	hydraURL   *url.URL
	retryOpts  []retry.Option
}

func NewTokenManager(client *http.Client, hydraURL *url.URL, retryOpts []retry.Option) *tokenManager {
	return &tokenManager{
		httpClient: client,
		hydraURL:   hydraURL.ResolveReference(&url.URL{Path: tokenEndpoint}),
		retryOpts:  retryOpts,
	}
}

func (t *tokenManager) FetchOAuth2Token(tokenRequest *TokenRequest) (*Token, error) {

	req, err := t.prepareHTTPRequest(tokenRequest)
	if err != nil {
		return nil, err
	}

	tokenResponse := &tokenResponse{}
	retryErr := retry.Do(func() error {

		resp, err := t.httpClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		if err = json.NewDecoder(resp.Body).Decode(tokenResponse); err != nil {
			return err
		}

		switch resp.StatusCode {
		case 200:
			return nil
		case 400, 401:
			return retry.Unrecoverable(errors.Wrap(tokenResponse.HydraError, "unrecoverable error, no retries"))
		default:
			return tokenResponse.HydraError
		}
	}, t.retryOpts...)

	if retryErr != nil {
		return nil, retryErr
	}

	return tokenResponse.Token, nil
}

func (t *tokenManager) prepareHTTPRequest(tr *TokenRequest) (*http.Request, error) {

	form := url.Values{}
	form.Set("grant_type", tr.GrantType)
	form.Set("scope", tr.Scope)

	req, err := http.NewRequest(http.MethodPost, t.hydraURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(tr.OAuth2ClientID, tr.OAuth2ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	return req, nil
}
