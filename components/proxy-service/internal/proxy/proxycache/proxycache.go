package proxycache

import (
	"net/http/httputil"
	"time"
	"github.com/patrickmn/go-cache"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"net/http"
)

const cleanupInterval = 60

type OauthCredentials struct {
	AuthenticationUrl string
	ClientId          string
	ClientSecret      string
}

type BasicAuthCredentials struct {
	UserName string
	Password string
}

type AuthorizationStrategy interface {
	Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError
	Reset()
}

type Credentials struct {
	Oauth *OauthCredentials
	Basic *BasicAuthCredentials
}

type Proxy struct {
	Proxy                 *httputil.ReverseProxy
	Credentials           *Credentials
	AuthorizationStrategy AuthorizationStrategy
}

type authorizationType int

const (
	None authorizationType = iota
	Basic
	OAuth
	Unknown
)

type HTTPProxyCache interface {
	Get(id string) (*Proxy, bool)
	Add(id, oauthUrl, clientId, clientSecret string, proxy *httputil.ReverseProxy) *Proxy
	PutWithCredentials(id string, credentials *Credentials, reverseProxy *httputil.ReverseProxy) *Proxy
}

type httpProxyCache struct {
	skipVerify bool
	proxyCache *cache.Cache
}

func NewProxyCache(skipVerify bool, proxyCacheTTL int) HTTPProxyCache {
	return &httpProxyCache{
		skipVerify: skipVerify,
		proxyCache: cache.New(time.Duration(proxyCacheTTL)*time.Second, cleanupInterval*time.Second),
	}
}

func (p *httpProxyCache) Get(id string) (*Proxy, bool) {
	proxy, found := p.proxyCache.Get(id)
	if !found {
		return nil, false
	}

	return proxy.(*Proxy), found
}

func (p *httpProxyCache) Add(id, oauthUrl, clientId, clientSecret string, reverseProxy *httputil.ReverseProxy) *Proxy {
	credentials := &Credentials{
		Oauth: &OauthCredentials{
			ClientId:          clientId,
			ClientSecret:      clientSecret,
			AuthenticationUrl: oauthUrl,
		},
	}
	proxy := &Proxy{Proxy: reverseProxy, Credentials: credentials}
	p.proxyCache.Set(id, proxy, cache.DefaultExpiration)

	return proxy
}

func (p *httpProxyCache) PutWithCredentials(id string, credentials *Credentials, reverseProxy *httputil.ReverseProxy) *Proxy {
	proxy := &Proxy{Proxy: reverseProxy, Credentials: credentials}
	p.proxyCache.Set(id, proxy, cache.DefaultExpiration)

	return proxy
}

func credentialsType(credentials *Credentials) authorizationType{
	if credentials == nil {
		return None
	} else if credentials.Basic != nil {
		return Basic
	} else if credentials.Oauth != nil {
		return OAuth
	} else {
		return Unknown
	}
}


