package client

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	log "github.com/sirupsen/logrus"
)

func New(timeoutDuration int, tokenCache TokenCache, httpClient *http.Client) csrf.Client {
	return &client{
		timeoutDuration: timeoutDuration,
		tokenCache:      tokenCache,
		httpClient:      httpClient,
	}
}

type client struct {
	timeoutDuration int
	tokenCache      TokenCache
	httpClient      *http.Client
}

func (c *client) GetTokenEndpointResponse(tokenEndpointURL string, strategy authorization.Strategy) (*csrf.Response, apperrors.AppError) {

	resp, found := c.tokenCache.Get(tokenEndpointURL)
	if found {
		return resp, nil
	}

	log.Infof("CSRF Token not found in cache, fetching (Endpoint: %s)", tokenEndpointURL)
	tokenResponse, err := c.requestToken(tokenEndpointURL, strategy, c.timeoutDuration)
	if err != nil {
		return nil, err
	}

	c.tokenCache.Add(tokenEndpointURL, tokenResponse)

	return tokenResponse, nil

}

func (c *client) InvalidateTokenCache(tokenEndpointURL string) {
	log.Infof("Invalidating token for endpoint: %s", tokenEndpointURL)
	c.tokenCache.Remove(tokenEndpointURL)
}

func (c *client) requestToken(csrfEndpointURL string, strategy authorization.Strategy, timeoutDuration int) (*csrf.Response, apperrors.AppError) {

	tokenRequest, err := http.NewRequest(http.MethodGet, csrfEndpointURL, strings.NewReader(""))
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	err = addAuthorization(tokenRequest, c.httpClient, strategy)
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	setCSRFSpecificHeaders(tokenRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := tokenRequest.WithContext(ctx)

	resp, err := c.httpClient.Do(requestWithContext)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to make a request to '%s': %s", csrfEndpointURL, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("incorrect response code '%d' while getting token from %s", resp.StatusCode, csrfEndpointURL)
	}

	tokenRes := &csrf.Response{
		CSRFToken: resp.Header.Get(httpconsts.HeaderCSRFToken),
		Cookies:   resp.Cookies(),
	}

	return tokenRes, nil

}

func addAuthorization(r *http.Request, client *http.Client, strategy authorization.Strategy) apperrors.AppError {
	return strategy.AddAuthorization(r, func(transport *http.Transport) {
		client.Transport = transport
	})
}

func setCSRFSpecificHeaders(r *http.Request) {
	r.Header.Add(httpconsts.HeaderCSRFToken, httpconsts.HeaderCSRFTokenVal)
	r.Header.Add(httpconsts.HeaderAccept, httpconsts.HeaderAcceptVal)
	r.Header.Add(httpconsts.HeaderCacheControl, httpconsts.HeaderCacheControlVal)
}
