package csrf

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
)

type Response struct {
	csrfToken string
	cookies   []*http.Cookie
}

type Client interface {
	GetTokenEndpointResponse(csrfEndpointURL string, strategy authorization.Strategy) (*Response, apperrors.AppError)
	InvalidateTokenCache(csrfEndpointURL string)
}

type client struct {
	timeoutDuration int
	tokenCache      TokenCache
}

func NewCSRFClient(timeoutDuration int, tokenCache TokenCache) Client {
	return &client{
		timeoutDuration: timeoutDuration,
		tokenCache:      tokenCache,
	}
}

func (c *client) GetTokenEndpointResponse(csrfEndpointURL string, strategy authorization.Strategy) (*Response, apperrors.AppError) {

	resp, found := c.tokenCache.Get(csrfEndpointURL)
	if found {
		//TODO: DEBUG
		log.Printf("Found cached Token Response: %#v", resp)
		return resp, nil
	}

	tokenResponse, err := c.requestToken(csrfEndpointURL, strategy)
	if err != nil {
		return nil, err
	}

	c.tokenCache.Add(csrfEndpointURL, tokenResponse)

	return tokenResponse, nil

}

func (c *client) InvalidateTokenCache(csrfEndpointURL string) {
	c.tokenCache.Remove(csrfEndpointURL)
}

func (c *client) requestToken(csrfEndpointURL string, strategy authorization.Strategy) (*Response, apperrors.AppError) {

	//TODO: DEBUG
	log.Printf("requestToken: csrfEndpointURL=%s", csrfEndpointURL)

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutDuration)*time.Second)
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
		csrfToken: resp.Header.Get(httpconsts.HeaderCSRFToken),
		cookies:   resp.Cookies(),
	}

	//TODO: DEBUG
	log.Printf("Token Response: %#v", tokenRes)

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
