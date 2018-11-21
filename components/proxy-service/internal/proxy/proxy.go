package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authorization"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata"
	metadatamodel "github.com/kyma-project/kyma/components/proxy-service/internal/metadata/model"
)

type proxy struct {
	nameResolver                 k8sconsts.NameResolver
	serviceDefService            metadata.ServiceDefinitionService
	cache                        Cache
	skipVerify                   bool
	proxyTimeout                 int
	authorizationStrategyFactory authorization.StrategyFactory
}

type Config struct {
	SkipVerify        bool
	ProxyTimeout      int
	RemoteEnvironment string
	ProxyCacheTTL     int
}

// New creates proxy for handling user's services calls
func New(serviceDefService metadata.ServiceDefinitionService, authorizationStrategyFactory authorization.StrategyFactory, config Config) http.Handler {
	return &proxy{
		nameResolver:                 k8sconsts.NewNameResolver(config.RemoteEnvironment),
		serviceDefService:            serviceDefService,
		cache:                        NewCache(config.ProxyCacheTTL),
		skipVerify:                   config.SkipVerify,
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
	}
}

// NewInvalidStateHandler creates handler always returning 500 response
func NewInvalidStateHandler(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleErrors(w, apperrors.Internal(message))
	})
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := p.extractServiceId(r.Host)

	cacheEntry, err := p.getOrCreateCacheEntry(id)
	if err != nil {
		handleErrors(w, err)
		return
	}

	newRequest, cancel := p.prepareRequest(r, cacheEntry)
	defer cancel()

	err = p.addAuthorizationHeader(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	p.addModifyResponseHandler(newRequest, id, cacheEntry)

	p.executeRequest(w, newRequest, cacheEntry)
}

func (p *proxy) extractServiceId(host string) string {
	return p.nameResolver.ExtractServiceId(host)
}

func (p *proxy) getOrCreateCacheEntry(id string) (*CacheEntry, apperrors.AppError) {
	cacheObj, found := p.cache.Get(id)

	if found {
		return cacheObj, nil
	} else {
		return p.createCacheEntry(id)
	}
}

func (p *proxy) createCacheEntry(id string) (*CacheEntry, apperrors.AppError) {
	serviceApi, err := p.serviceDefService.GetAPI(id)
	if err != nil {
		return nil, err
	}

	proxy, err := makeProxy(serviceApi.TargetUrl, id, p.skipVerify)
	if err != nil {
		return nil, err
	}

	authorizationStrategy := p.newAuthorizationStrategy(serviceApi.Credentials)

	return p.cache.Put(id, proxy, authorizationStrategy), nil
}

func (p *proxy) newAuthorizationStrategy(credentials *metadatamodel.Credentials) authorization.Strategy {
	return p.authorizationStrategyFactory.Create(credentials)
}

func (p *proxy) prepareRequest(r *http.Request, cacheEntry *CacheEntry) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.proxyTimeout)*time.Second)
	newRequest := r.WithContext(ctx)

	return newRequest, cancel
}

func (p *proxy) addAuthorizationHeader(r *http.Request, cacheEntry *CacheEntry) apperrors.AppError {
	return cacheEntry.AuthorizationStrategy.AddAuthorizationHeader(r)
}

func (p *proxy) addModifyResponseHandler(r *http.Request, id string, cacheEntry *CacheEntry) {
	cacheEntry.Proxy.ModifyResponse = p.createModifyResponseFunction(id, r)
}

func (p *proxy) createModifyResponseFunction(id string, r *http.Request) func(*http.Response) error {
	// Handle the case when credentials has been changed or OAuth token has expired
	return func(response *http.Response) error {
		retrier := newUnathorizedResponseRetrier(id, r, p.proxyTimeout, p.createCacheEntry)

		return retrier.RetryIfFailedToAuthorize(response)
	}
}

func (p *proxy) executeRequest(w http.ResponseWriter, r *http.Request, cacheEntry *CacheEntry) {
	cacheEntry.Proxy.ServeHTTP(w, r)
}

func handleErrors(w http.ResponseWriter, apperr apperrors.AppError) {
	code, body := httperrors.AppErrorToResponse(apperr)
	respondWithBody(w, code, body)
}

func respondWithBody(w http.ResponseWriter, code int, body httperrors.ErrorResponse) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)

	w.WriteHeader(code)

	json.NewEncoder(w).Encode(body)
}
