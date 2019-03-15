package csrf

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	CSRFToken string
	Cookies   []*http.Cookie
}

type Client interface {
	GetTokenEndpointResponse(csrfEndpointURL string, strategy authorization.Strategy) (*Response, apperrors.AppError)
	InvalidateTokenCache(csrfEndpointURL string)
}

type client struct {
	timeoutDuration int
	tokenCache      TokenCache
	requestToken    func(csrfEndpointURL string, strategy authorization.Strategy, timeoutDuration int) (*Response, apperrors.AppError)
}

func NewCSRFClient(timeoutDuration int, tokenCache TokenCache) Client {
	return &client{
		timeoutDuration: timeoutDuration,
		tokenCache:      tokenCache,
		requestToken:    requestToken,
	}
}

func (c *client) GetTokenEndpointResponse(tokenEndpointURL string, strategy authorization.Strategy) (*Response, apperrors.AppError) {

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

func requestToken(csrfEndpointURL string, strategy authorization.Strategy, timeoutDuration int) (*Response, apperrors.AppError) {

	client := &http.Client{}

	tokenRequest, err := http.NewRequest(http.MethodGet, csrfEndpointURL, strings.NewReader(""))
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	err = addAuthorization(tokenRequest, client, strategy)
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	setCSRFSpecificHeaders(tokenRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := tokenRequest.WithContext(ctx)

	resp, err := client.Do(requestWithContext)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("failed to make a request to '%s': %s", csrfEndpointURL, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("incorrect response code '%d' while getting token from %s", resp.StatusCode, csrfEndpointURL)
	}

	tokenRes := &Response{
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
