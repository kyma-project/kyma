package client

import (
	"context"
	"crypto/tls"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
)

func New(timeoutDuration int, tokenCache TokenCache) csrf.Client {
	clientCertificate := clientcert.NewClientCertificate(nil)

	return &client{
		timeoutDuration:   timeoutDuration,
		tokenCache:        tokenCache,
		clientCertificate: clientCertificate,
	}
}

type client struct {
	timeoutDuration   int
	tokenCache        TokenCache
	clientCertificate clientcert.ClientCertificate
}

func (c *client) GetTokenEndpointResponse(tokenEndpointURL string, strategy authorization.Strategy, skipTLSVerify bool) (*csrf.Response, apperrors.AppError) {

	resp, found := c.tokenCache.Get(tokenEndpointURL)
	if found {
		return resp, nil
	}

	zap.L().Info("CSRF Token not found in cache, fetching",
		zap.String("tokenEndpoint", tokenEndpointURL))

	tokenResponse, err := c.requestToken(tokenEndpointURL, strategy, c.timeoutDuration, skipTLSVerify)
	if err != nil {
		return nil, err
	}

	c.tokenCache.Add(tokenEndpointURL, tokenResponse)

	return tokenResponse, nil

}

func (c *client) InvalidateTokenCache(tokenEndpointURL string) {
	zap.L().Info("Invalidating token for endpoint",
		zap.String("tokenEndpoint", tokenEndpointURL))
	c.tokenCache.Remove(tokenEndpointURL)
}

func (c *client) requestToken(csrfEndpointURL string, strategy authorization.Strategy, timeoutDuration int, skipTLSVerify bool) (*csrf.Response, apperrors.AppError) {

	tokenRequest, err := http.NewRequest(http.MethodGet, csrfEndpointURL, strings.NewReader(""))
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	err = addAuthorization(tokenRequest, c.clientCertificate, strategy, skipTLSVerify)
	if err != nil {
		return nil, apperrors.Internal("failed to create token request: %s", err.Error())
	}

	setCSRFSpecificHeaders(tokenRequest)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutDuration)*time.Second)
	defer cancel()
	requestWithContext := tokenRequest.WithContext(ctx)

	httpClient := &http.Client{
		Transport: httptools.NewRoundTripper(httptools.WithGetClientCertificate(c.clientCertificate.GetClientCertificate), httptools.WithTLSSkipVerify(skipTLSVerify)),
	}
	resp, err := httpClient.Do(requestWithContext)
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

func addAuthorization(r *http.Request, clientCertificate clientcert.ClientCertificate, strategy authorization.Strategy, skipTLSVerify bool) apperrors.AppError {
	return strategy.AddAuthorization(r, func(cert *tls.Certificate) {
		clientCertificate.SetCertificate(cert)
	}, skipTLSVerify)
}

func setCSRFSpecificHeaders(r *http.Request) {
	r.Header.Add(httpconsts.HeaderCSRFToken, httpconsts.HeaderCSRFTokenVal)
	r.Header.Add(httpconsts.HeaderAccept, httpconsts.HeaderAcceptVal)
	r.Header.Add(httpconsts.HeaderCacheControl, httpconsts.HeaderCacheControlVal)
}
