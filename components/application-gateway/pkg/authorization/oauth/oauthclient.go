package oauth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/util"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization/oauth/tokencache"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
)

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type Client interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateAndRetry(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateTokenCache(clientID string)
}

type client struct {
	timeoutDuration int
	tokenCache      tokencache.TokenCache
}

func NewOauthClient(timeoutDuration int, tokenCache tokencache.TokenCache) Client {
	return &client{
		timeoutDuration: timeoutDuration,
		tokenCache:      tokenCache,
	}
}

func (c *client) GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError) {
	token, found := c.tokenCache.Get(clientID)
	if found {
		return token, nil
	}

	tokenResponse, err := c.requestToken(clientID, clientSecret, authURL)
	if err != nil {
		return "", err
	}

	c.tokenCache.Add(clientID, tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return tokenResponse.AccessToken, nil
}

func (c *client) InvalidateAndRetry(clientID string, clientSecret string, authURL string) (string, apperrors.AppError) {
	c.tokenCache.Remove(clientID)

	tokenResponse, err := c.requestToken(clientID, clientSecret, authURL)
	if err != nil {
		return "", err
	}

	c.tokenCache.Add(clientID, tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return tokenResponse.AccessToken, nil
}

func (c *client) InvalidateTokenCache(clientID string) {
	c.tokenCache.Remove(clientID)
}

func (c *client) requestToken(clientID string, clientSecret string, authURL string) (*oauthResponse, apperrors.AppError) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}

	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("client_secret", clientSecret)
	form.Add("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	util.AddBasicAuthHeader(req, clientID, clientSecret)
	req.Header.Add(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationURLEncoded)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := req.WithContext(ctx)

	response, err := client.Do(requestWithContext)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to make a request to '%s': %s", authURL, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("incorrect response code '%d' while getting token from %s", response.StatusCode, authURL)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to read token response body from '%s': %s", authURL, err.Error())
	}

	tokenResponse := &oauthResponse{}

	err = json.Unmarshal(body, tokenResponse)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to unmarshal token response body: %s", err.Error())
	}

	return tokenResponse, nil
}
