package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
)

func makeProxy(targetURL string, requestParameters *authorization.RequestParameters, serviceName string, skipTLSVerify bool, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy, clientCertificate clientcert.ClientCertificate, timeout int) (*httputil.ReverseProxy, apperrors.AppError) {
	roundTripper := httptools.NewRoundTripper(httptools.WithTLSSkipVerify(skipTLSVerify), httptools.WithGetClientCertificate(clientCertificate.GetClientCertificate))
	retryableRoundTripper := NewRetryableRoundTripper(roundTripper, authorizationStrategy, csrfTokenStrategy, clientCertificate, timeout, skipTLSVerify)
	return newProxy(targetURL, requestParameters, serviceName, retryableRoundTripper)
}

func newProxy(targetURL string, requestParameters *authorization.RequestParameters, serviceName string, transport http.RoundTripper) (*httputil.ReverseProxy, apperrors.AppError) {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Errorf("failed to parse target url '%s': '%s'", targetURL, err.Error())
		return nil, apperrors.Internal("failed to parse target url '%s': '%s'", targetURL, err.Error())
	}

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		log.Infof("Proxy call for service '%s' to '%s'", serviceName, targetURL)

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		combinedPath := joinPaths(target.Path, req.URL.Path)
		combinedPathEscaped := joinPaths(target.Path, req.URL.RawPath)
		req.URL.Path = combinedPath
		req.URL.RawPath = combinedPathEscaped

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		if requestParameters != nil {
			setCustomQueryParameters(req.URL, requestParameters.QueryParameters)
			setCustomHeaders(req.Header, requestParameters.Headers)
		}

		log.Infof("Modified request url : '%s', schema : '%s', path : '%s'", req.URL.String(), req.URL.Scheme, req.URL.Path)
	}
	errorHandler := func(rw http.ResponseWriter, req *http.Request, err error) {
		codeRewriter(rw, err)
	}
	return &httputil.ReverseProxy{Director: director, Transport: transport, ErrorHandler: errorHandler}, nil
}

func joinPaths(a, b string) string {
	if b == "" {
		return a
	}

	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
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

func responseModifier(
	gatewayURL *url.URL,
	targetURL string,
	urlRewriter func(gatewayURL, target, loc *url.URL) *url.URL,
) func(*http.Response) error {
	return func(resp *http.Response) error {
		if (resp.StatusCode < 300 || resp.StatusCode >= 400) &&
			resp.StatusCode != http.StatusCreated {
			return nil
		}

		const locationHeader = "Location"

		locRaw := resp.Header.Get(locationHeader)

		if locRaw == "" {
			return nil
		}

		loc, err := resp.Request.URL.Parse(locRaw)
		if err != nil {
			return nil
		}

		target, err := url.Parse(targetURL)
		if err != nil {
			return nil
		}

		newURL := urlRewriter(gatewayURL, target, loc)

		if newURL != nil {
			resp.Header.Set(locationHeader, newURL.String())
		}

		return nil
	}
}

// urlRewriter modifies redirect URLs for reverse proxy.
// If the URL should be left unmodified - it returns nil.
func urlRewriter(gatewayURL, target, loc *url.URL) *url.URL {
	if loc.Scheme != "http" && loc.Scheme != "https" {
		return nil
	}

	if loc.Hostname() != target.Hostname() || !strings.HasPrefix(loc.Path, target.Path) {
		return nil
	}

	stripped := strings.TrimPrefix(loc.Path, target.Path)
	gatewayURL = gatewayURL.JoinPath(stripped)
	gatewayURL.RawQuery = loc.RawQuery
	gatewayURL.Fragment = loc.Fragment

	return gatewayURL
}

func codeRewriter(rw http.ResponseWriter, err error) {

	if errors.Is(err, context.DeadlineExceeded) {
		log.Infof("%s: HTTP status code was rewritten to 504", err)
		rw.WriteHeader(http.StatusGatewayTimeout)
		return
	}
	rw.WriteHeader(http.StatusBadGateway)
}
