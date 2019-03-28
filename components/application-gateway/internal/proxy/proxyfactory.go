package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/application-gateway/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/httpconsts"
	log "github.com/sirupsen/logrus"
)

func makeProxy(targetUrl string, id string, skipVerify bool) (*httputil.ReverseProxy, apperrors.AppError) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		log.Errorf("failed to parse target url '%s': '%s'", targetUrl, err.Error())
		return nil, apperrors.Internal("failed to parse target url '%s': '%s'", targetUrl, err.Error())
	}

	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		log.Infof("Proxy call for service '%s' to '%s'", id, targetUrl)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host

		req.URL.Path = joinPaths(target.Path, req.URL.Path)

		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}

		if _, ok := req.Header[httpconsts.HeaderUserAgent]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set(httpconsts.HeaderUserAgent, "")
		}

		removeHeader(req.Header, httpconsts.HeaderXForwardedProto)
		removeHeader(req.Header, httpconsts.HeaderXForwardedFor)
		removeHeader(req.Header, httpconsts.HeaderXForwardedHost)
		removeHeader(req.Header, httpconsts.HeaderXForwardedClientCert)

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

func removeHeader(headers http.Header, headerToRemove string) {
	if _, ok := headers[headerToRemove]; ok {
		log.Debugf("Removing header %s\n", headerToRemove)
		headers.Del(headerToRemove)
	}
}
