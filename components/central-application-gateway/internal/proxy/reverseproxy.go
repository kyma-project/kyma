package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httptools"
	log "github.com/sirupsen/logrus"
)

func makeProxy(targetURL string, requestParameters *authorization.RequestParameters, serviceName string, skipVerify bool) (*httputil.ReverseProxy, apperrors.AppError) {
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
		req.URL.RawPath = combinedPath
		req.URL.Path = combinedPath

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		if requestParameters != nil {
			setCustomQueryParameters(req.URL, requestParameters.QueryParameters)
			setCustomHeaders(req.Header, requestParameters.Headers)
		}

		removeForbiddenHeaders(req.Header)

		log.Infof("Modified request url : '%s', schema : '%s', path : '%s'", req.URL.String(), req.URL.Scheme, req.URL.Path)
	}
	newProxy := &httputil.ReverseProxy{Director: director}

	newProxy.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify}}

	return newProxy, nil
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

func removeForbiddenHeaders(reqHeaders http.Header) {
	httptools.RemoveHeader(reqHeaders, httpconsts.HeaderXForwardedProto)
	httptools.RemoveHeader(reqHeaders, httpconsts.HeaderXForwardedFor)
	httptools.RemoveHeader(reqHeaders, httpconsts.HeaderXForwardedHost)
	httptools.RemoveHeader(reqHeaders, httpconsts.HeaderXForwardedClientCert)
}

func setCustomHeaders(reqHeaders http.Header, customHeaders *map[string][]string) {
	if _, ok := reqHeaders[httpconsts.HeaderUserAgent]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		reqHeaders.Set(httpconsts.HeaderUserAgent, "")
	}

	httptools.SetHeaders(reqHeaders, customHeaders)
}
