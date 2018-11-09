package proxy

import (
	gocache "github.com/patrickmn/go-cache"
	"net/http/httputil"
	"time"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization"
)

const cleanupInterval = 60

type CacheEntry struct {
	Proxy                 *httputil.ReverseProxy
	AuthorizationStrategy authorization.Strategy
}

type Cache interface {
	// Get returns entry from the cache
	Get(id string) (*CacheEntry, bool)
	// Put adds entry to the cache
	Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy) *CacheEntry
}

type cache struct {
	proxyCache *gocache.Cache
}

// NewCache creates new cache with specified TTL
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

func (p *cache) Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy) *CacheEntry {
	proxy := &CacheEntry{Proxy: reverseProxy, AuthorizationStrategy: authorizationStrategy}
	p.proxyCache.Set(id, proxy, gocache.DefaultExpiration)

	return proxy
}
