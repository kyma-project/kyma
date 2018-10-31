package proxycache

import (
	"net/http/httputil"
	"time"

	"github.com/patrickmn/go-cache"
	"net/http"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
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

type Credentials struct {
	Type  string
	Oauth *OauthCredentials
	Basic *BasicAuthCredentials
}

type AuthorizationStrategy interface {
	Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError
	Reset()
}

type RetryStrategy interface {
	Do(r *http.Response) apperrors.AppError
}


type Proxy struct {
	Proxy                 *httputil.ReverseProxy
	AuthorizationStrategy AuthorizationStrategy
	Credentials           *Credentials
}

type HTTPProxyCache interface {
	Get(id string) (*Proxy, bool)
	Add(id, oauthUrl, clientId, clientSecret string, proxy *httputil.ReverseProxy) *Proxy
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
