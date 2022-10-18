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

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/tokencache"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/util"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
)

type oauthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

//go:generate mockery --name=Client
type Client interface {
	GetToken(clientID, clientSecret, authURL string, headers, queryParameters *map[string][]string, skipVerify bool) (string, apperrors.AppError)
	GetTokenMTLS(clientID, clientSecret string, authURL string, cert tls.Certificate, headers, queryParameters *map[string][]string, skipVerify bool) (string, apperrors.AppError)
	InvalidateTokenCache(clientID string, clientSecret string, authURL string)
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

func (c *client) GetToken(clientID, clientSecret, authURL string, headers, queryParameters *map[string][]string, skipVerify bool) (string, apperrors.AppError) {
	token, found := c.tokenCache.Get(c.makeOAuthTokenCacheKey(clientID, clientSecret, authURL))
	if found {
		return token, nil
	}

	tokenResponse, err := c.requestToken(clientID, clientSecret, authURL, headers, queryParameters, skipVerify)
	if err != nil {
		return "", err
	}

	c.tokenCache.Add(c.makeOAuthTokenCacheKey(clientID, clientSecret, authURL), tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return tokenResponse.AccessToken, nil
}

func (c *client) GetTokenMTLS(clientID, clientSecret string, authURL string, cert tls.Certificate, headers, queryParameters *map[string][]string, skipVerify bool) (string, apperrors.AppError) {
	token, found := c.tokenCache.Get(c.makeOAuthTokenCacheKey(clientID, clientSecret, authURL))
	if found {
		return token, nil
	}

	tokenResponse, err := c.requestTokenMTLS(clientID, authURL, cert, headers, queryParameters, skipVerify)
	if err != nil {
		return "", err
	}

	c.tokenCache.Add(c.makeOAuthTokenCacheKey(clientID, clientSecret, authURL), tokenResponse.AccessToken, tokenResponse.ExpiresIn)

	return tokenResponse.AccessToken, nil
}

func (c *client) InvalidateTokenCache(clientID, clientSecret, authURL string) {
	c.tokenCache.Remove(c.makeOAuthTokenCacheKey(clientID, clientSecret, authURL))
}

// to avoid case of single clientID and different endpoints for MTLS and standard oauth
func (c *client) makeOAuthTokenCacheKey(clientID, clientSecret, authURL string) string {
	return clientID + clientSecret + authURL
}

func (c *client) requestToken(clientID, clientSecret, authURL string, headers, queryParameters *map[string][]string, skipVerify bool) (*oauthResponse, apperrors.AppError) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
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

	setCustomQueryParameters(req.URL, queryParameters)
	setCustomHeaders(req.Header, headers)

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

func (c *client) requestTokenMTLS(clientID, authURL string, cert tls.Certificate, headers, queryParameters *map[string][]string, skipVerify bool) (*oauthResponse, apperrors.AppError) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: skipVerify,
		},
	}
	client := &http.Client{Transport: transport}

	form := url.Values{}
	form.Add("client_id", clientID)
	form.Add("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, authURL, strings.NewReader(form.Encode()))

	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	req.Header.Add(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationURLEncoded)

	setCustomQueryParameters(req.URL, queryParameters)
	setCustomHeaders(req.Header, headers)

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

func setCustomQueryParameters(reqURL *url.URL, customQueryParams *map[string][]string) {
	httptools.SetQueryParameters(reqURL, customQueryParams)
}

func setCustomHeaders(reqHeaders http.Header, customHeaders *map[string][]string) {
	if _, ok := reqHeaders[httpconsts.HeaderUserAgent]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		reqHeaders.Set(httpconsts.HeaderUserAgent, "")
	}

	httptools.SetHeaders(reqHeaders, customHeaders)
}
