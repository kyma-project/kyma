package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	gocache "github.com/patrickmn/go-cache"
)

const cleanupInterval = 60

// CacheEntry stores information about proxy configuration in cache
type CacheEntry struct {
	Proxy                 *httputil.ReverseProxy
	AuthorizationStrategy *authorizationStrategyWrapper
	CSRFTokenStrategy     csrf.TokenStrategy
}

type authorizationStrategyWrapper struct {
	actualStrategy    authorization.Strategy
	proxy             *httputil.ReverseProxy
	clientCertificate clientcert.ClientCertificate
}

func (ce *authorizationStrategyWrapper) AddAuthorization(r *http.Request) apperrors.AppError {
	return ce.actualStrategy.AddAuthorization(r, ce.clientCertificate.SetCertificate)
}

func (ce *authorizationStrategyWrapper) Invalidate() {
	ce.actualStrategy.Invalidate()
}

// Cache is an interface for caching Proxies
type Cache interface {
	// Get returns entry from the cache
	Get(appName, serviceName, apiName string) (*CacheEntry, bool)
	// Put adds entry to the cache
	Put(appName, serviceName, apiName string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy, clientCertificate clientcert.ClientCertificate) *CacheEntry
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

func (p *cache) Get(appName, serviceName, apiName string) (*CacheEntry, bool) {
	key := appName + serviceName + apiName
	proxy, found := p.proxyCache.Get(key)
	if !found {
		return nil, false
	}

	return proxy.(*CacheEntry), found
}

func (p *cache) Put(appName, serviceName, apiName string, reverseProxy *httputil.ReverseProxy, authorizationStrategy authorization.Strategy, csrfTokenStrategy csrf.TokenStrategy, clientCertificate clientcert.ClientCertificate) *CacheEntry {
	key := appName + serviceName + apiName
	proxy := &CacheEntry{Proxy: reverseProxy, AuthorizationStrategy: &authorizationStrategyWrapper{authorizationStrategy, reverseProxy, clientCertificate}, CSRFTokenStrategy: csrfTokenStrategy}
	p.proxyCache.Set(key, proxy, gocache.DefaultExpiration)

	return proxy
}
