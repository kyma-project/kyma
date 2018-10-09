package proxycache

import (
	"net/http/httputil"
	"time"

	"github.com/patrickmn/go-cache"
)

const cleanupInterval = 60

type Proxy struct {
	Proxy        *httputil.ReverseProxy
	OauthUrl     string
	ClientId     string
	ClientSecret string
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
	proxy := &Proxy{Proxy: reverseProxy, OauthUrl: oauthUrl, ClientId: clientId, ClientSecret: clientSecret}
	p.proxyCache.Set(id, proxy, cache.DefaultExpiration)

	return proxy
}
