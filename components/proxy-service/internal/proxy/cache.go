package proxy

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	gocache "github.com/patrickmn/go-cache"
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

type Cache interface {
	Get(id string) (*CacheEntry, bool)
	Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy AuthorizationStrategy) *CacheEntry
}

type cache struct {
	proxyCache *gocache.Cache
}

func NewCache(proxyCacheTTL int) Cache {
	return &cache{
		proxyCache: gocache.New(time.Duration(proxyCacheTTL)*time.Second, cleanupInterval*time.Second),
	}
}

func (p *cache) Get(id string) (*CacheEntry, bool) {
	proxy, found := p.proxyCache.Get(id)
	if !found {
		return nil, false
	}

	return proxy.(*CacheEntry), found
}

func (p *cache) Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy AuthorizationStrategy) *CacheEntry {
	proxy := &CacheEntry{Proxy: reverseProxy, AuthorizationStrategy: authorizationStrategy}
	p.proxyCache.Set(id, proxy, gocache.DefaultExpiration)

	return proxy
}
