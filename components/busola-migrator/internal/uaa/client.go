package uaa

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
)

const (
	wellKnown = "/.well-known/openid-configuration"
)

type Client struct {
	uaaConfig config.UAAConfig

	httpClient *http.Client
}

func NewClient(uaaConfig config.UAAConfig) Client {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	return Client{
		httpClient: httpClient,
		uaaConfig:  uaaConfig,
	}
}

type OpenIDConfiguration struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

func (c Client) GetOpenIDConfiguration() (OpenIDConfiguration, error) {
	res, err := c.httpClient.Get(c.uaaConfig.URL + wellKnown)
	if err != nil {
		return OpenIDConfiguration{}, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return OpenIDConfiguration{}, err
	}

	var oidcConfig OpenIDConfiguration
	err = json.Unmarshal(body, &oidcConfig)
	if err != nil {
		return OpenIDConfiguration{}, err
	}

	return oidcConfig, nil
}

func (c Client) GetAuthorizationEndpointWithParams(authzEndpoint, oauthState string) (string, error) {
	authURL, err := url.Parse(authzEndpoint)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("client_id", c.uaaConfig.ClientID)
	params.Add("redirect_uri", c.uaaConfig.RedirectURI)
	params.Add("response_type", "code")
	params.Add("state", oauthState)

	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

func (c Client) GetToken(tokenEndpoint string, authCode string) (map[string]interface{}, error) {
	params := url.Values{
		"redirect_uri": {c.uaaConfig.RedirectURI},
		"grant_type":   {"authorization_code"},
		"code":         {authCode},
	}

	req, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.uaaConfig.ClientID, c.uaaConfig.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resJSON map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&resJSON)
	if err != nil {
		return nil, err
	}

	return resJSON, nil
}
