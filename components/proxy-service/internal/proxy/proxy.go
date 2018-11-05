package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/authentication"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/proxycache"
)

type proxy struct {
	nameResolver                  k8sconsts.NameResolver
	serviceDefService             metadata.ServiceDefinitionService
	httpProxyCache                proxycache.HTTPProxyCache
	skipVerify                    bool
	proxyTimeout                  int
	authenticationStrategyFactory authentication.StrategyFactory
}

type Config struct {
	SkipVerify        bool
	ProxyTimeout      int
	Namespace         string
	RemoteEnvironment string
	ProxyCacheTTL     int
}

// New creates proxy for handling user's services calls
func New(serviceDefService metadata.ServiceDefinitionService, authenticationStrategyFactory authentication.StrategyFactory, config Config) http.Handler {
	return &proxy{
		nameResolver:                  k8sconsts.NewNameResolver(config.RemoteEnvironment, config.Namespace),
		serviceDefService:             serviceDefService,
		httpProxyCache:                proxycache.NewProxyCache(config.SkipVerify, config.ProxyCacheTTL),
		skipVerify:                    config.SkipVerify,
		proxyTimeout:                  config.ProxyTimeout,
		authenticationStrategyFactory: authenticationStrategyFactory,
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

	cacheEntry, err := p.getCacheEntry(id)
	if err != nil {
		handleErrors(w, err)
		return
	}

	newRequest, cancel := p.prepareRequest(r, cacheEntry)
	defer cancel()

	err = p.addAuthorization(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	cacheEntry.Proxy.ServeHTTP(w, newRequest)
}

func (p *proxy) extractServiceId(host string) string {
	return p.nameResolver.ExtractServiceId(host)
}

func (p *proxy) getCacheEntry(id string) (*proxycache.CacheEntry, apperrors.AppError) {
	cacheObj, found := p.httpProxyCache.Get(id)

	if found {
		return cacheObj, nil
	} else {
		return p.newCacheEntry(id)
	}
}

func (p *proxy) newCacheEntry(id string) (*proxycache.CacheEntry, apperrors.AppError) {
	serviceApi, err := p.serviceDefService.GetAPI(id)
	if err != nil {
		return nil, err
	}

	proxy, err := makeProxy(serviceApi.TargetUrl, id, p.skipVerify)
	if err != nil {
		return nil, err
	}

	authenticationStrategy := p.newAuthenticationStrategy(serviceApi.Credentials)

	return p.httpProxyCache.Put(id, proxy, authenticationStrategy), nil
}

func (p *proxy) prepareRequest(r *http.Request, cacheEntry *proxycache.CacheEntry) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.proxyTimeout)*time.Second)
	newRequest := r.WithContext(ctx)

	return newRequest, cancel
}

func (p *proxy) addAuthorization(r *http.Request, cacheEntry *proxycache.CacheEntry) apperrors.AppError {
	return cacheEntry.AuthorizationStrategy.Setup(r)
}

func (p *proxy) newAuthenticationStrategy(credentials *serviceapi.Credentials) authentication.Strategy {
	authCredentials := authentication.Credentials{}
	
	if oauthCredentialsProvided(credentials) {
		authCredentials = authentication.Credentials{
			Oauth: &authentication.OauthCredentials{
				ClientId:          credentials.Oauth.ClientID,
				ClientSecret:      credentials.Oauth.ClientSecret,
				AuthenticationUrl: credentials.Oauth.URL,
			},
		}
	}

	return p.authenticationStrategyFactory.Create(authCredentials)
}

func respondWithBody(w http.ResponseWriter, code int, body httperrors.ErrorResponse) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)

	w.WriteHeader(code)

	json.NewEncoder(w).Encode(body)
}

func handleErrors(w http.ResponseWriter, apperr apperrors.AppError) {
	code, body := httperrors.AppErrorToResponse(apperr)
	respondWithBody(w, code, body)
}

func oauthCredentialsProvided(credentials *serviceapi.Credentials) bool {
	return credentials != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}
