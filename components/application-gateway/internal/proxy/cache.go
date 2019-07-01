package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	gocache "github.com/patrickmn/go-cache"
)

const cleanupInterval = 60

type CacheEntry struct {
	Proxy                 *httputil.ReverseProxy
	AuthorizationStrategy *authorizationStrategyWrapper
	CSRFTokenStrategy     csrf.TokenStrategy
}

type authorizationStrategyWrapper struct {
	actualStrategy authorization.Strategy
	proxy          *httputil.ReverseProxy
}

func (ce *authorizationStrategyWrapper) AddAuthorization(r *http.Request) apperrors.AppError {

	ts := func(transport *http.Transport) {
		ce.proxy.Transport = transport
	}

	return ce.actualStrategy.AddAuthorization(r, ts)
}

func (ce *authorizationStrategyWrapper) Invalidate() {
	ce.actualStrategy.Invalidate()
}

type Cache interface {
	// Get returns entry from the cache
	Get(id string) (*CacheEntry, bool)
	// Put adds entry to the cache
	Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy) *CacheEntry
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

func (p *cache) Put(id string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy) *CacheEntry {

	proxy := &CacheEntry{Proxy: reverseProxy, AuthorizationStrategy: &authorizationStrategyWrapper{authorizationStrategy, reverseProxy}, CSRFTokenStrategy: csrfTokenStrategy}
	p.proxyCache.Set(id, proxy, gocache.DefaultExpiration)

	return proxy
}
