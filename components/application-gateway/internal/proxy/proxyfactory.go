package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httptools"
	log "github.com/sirupsen/logrus"
)

func makeProxy(targetUrl string, requestParameters *authorization.RequestParameters, id string, skipVerify bool) (*httputil.ReverseProxy, apperrors.AppError) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		log.Errorf("failed to parse target url '%s': '%s'", targetUrl, err.Error())
		return nil, apperrors.Internal("failed to parse target url '%s': '%s'", targetUrl, err.Error())
	}

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		log.Infof("Proxy call for service '%s' to '%s'", id, targetUrl)

		strippedPath := stripSecretFromPath(req.URL.Path)
		log.Infof("Striped strippedPath: %s", strippedPath)

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		combinedPath := joinPaths(target.Path, strippedPath)
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

func stripSecretFromPath(path string) string {
	segments := strings.Split(path, "/")
	if len(segments) < 3 || segments[1] != "secret" {
		return path
	}

	strippedPath := fmt.Sprintf("/%s", strings.Join(segments[3:], "/"))

	if !strings.HasSuffix(path, "/") {
		strippedPath = strings.TrimSuffix(strippedPath, "/")
	}

	return strippedPath
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
