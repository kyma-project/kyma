package proxycache

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/patrickmn/go-cache"
	"net/http"
	"net/http/httputil"
	"time"
)

const cleanupInterval = 60

type AuthorizationStrategy interface {
	Setup(r *http.Request) apperrors.AppError
	Reset()
}

type CacheEntry struct {
	Proxy                 *httputil.ReverseProxy
	AuthorizationStrategy AuthorizationStrategy
}

type HTTPProxyCache interface {
	Get(id string) (*CacheEntry, bool)
	Add(id, oauthUrl, clientId, clientSecret string, proxy *httputil.ReverseProxy) *CacheEntry
	Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy AuthorizationStrategy) *CacheEntry
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

func (p *httpProxyCache) Get(id string) (*CacheEntry, bool) {
	proxy, found := p.proxyCache.Get(id)
	if !found {
		return nil, false
	}

	return proxy.(*CacheEntry), found
}

func (p *httpProxyCache) Add(id, oauthUrl, clientId, clientSecret string, reverseProxy *httputil.ReverseProxy) *CacheEntry {

	proxy := &CacheEntry{Proxy: reverseProxy}
	p.proxyCache.Set(id, proxy, cache.DefaultExpiration)

	return proxy
}

func (p *httpProxyCache) Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy AuthorizationStrategy) *CacheEntry {
	proxy := &CacheEntry{Proxy: reverseProxy, AuthorizationStrategy: authorizationStrategy}
	p.proxyCache.Set(id, proxy, cache.DefaultExpiration)

	return proxy
}
