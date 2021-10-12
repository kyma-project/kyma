package proxy

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
)

type RetryableRoundTripper struct {
	roundTripper          http.RoundTripper
	authorizationStrategy authorization.Strategy
	csrfTokenStrategy     csrf.TokenStrategy
	clientCertificate     clientcert.ClientCertificate
	timeout               int
}

func NewRetryableRoundTripper(roundTripper http.RoundTripper, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy, clientCertificate clientcert.ClientCertificate, timeout int) *RetryableRoundTripper {
	return &RetryableRoundTripper{
		roundTripper:          roundTripper,
		authorizationStrategy: authorizationStrategy,
		csrfTokenStrategy:     csrfTokenStrategy,
		clientCertificate:     clientCertificate,
		timeout:               timeout,
	}
}

func (p *RetryableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Handle the case when credentials has been changed or OAuth token has expired
	secondRequestBody, copyErr := copyRequestBody(req)
	if copyErr != nil {
		return nil, copyErr
	}
	resp, err := p.roundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if !p.shouldRetry(resp) {
		return resp, err
	}
	if req.Context().Err() != nil {
		return nil, req.Context().Err()
	}
	return p.retry(req, secondRequestBody)
}

func (p *RetryableRoundTripper) shouldRetry(resp *http.Response) bool {
	return resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized
}

func (p *RetryableRoundTripper) retry(req *http.Request, retryBody io.ReadCloser) (*http.Response, error) {
	request, cancel := p.prepareRequest(req)
	defer cancel()
	request.Body = retryBody
	if err := p.addAuthorization(request); err != nil {
		return nil, err
	}

	return p.roundTripper.RoundTrip(request)
}

func (p *RetryableRoundTripper) prepareRequest(req *http.Request) (*http.Request, context.CancelFunc) {
	req.RequestURI = ""
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.timeout)*time.Second)
	return req.WithContext(ctx), cancel
}

func (p *RetryableRoundTripper) addAuthorization(r *http.Request) error {
	authorizationStrategy := p.authorizationStrategy
	authorizationStrategy.Invalidate()
	err := authorizationStrategy.AddAuthorization(r, p.clientCertificate.SetCertificate)
	if err != nil {
		return err
	}
	csrfTokenStrategy := p.csrfTokenStrategy
	csrfTokenStrategy.Invalidate()
	return csrfTokenStrategy.AddCSRFToken(r)
}
