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
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
)

type proxy struct {
	nameResolver                  k8sconsts.NameResolver
	serviceDefService             metadata.ServiceDefinitionService
	cache                         Cache
	skipVerify                    bool
	proxyTimeout                  int
	authenticationStrategyFactory authorization.StrategyFactory
}

type Config struct {
	SkipVerify        bool
	ProxyTimeout      int
	Namespace         string
	RemoteEnvironment string
	ProxyCacheTTL     int
}

// New creates proxy for handling user's services calls
func New(serviceDefService metadata.ServiceDefinitionService, authenticationStrategyFactory authorization.StrategyFactory, config Config) http.Handler {
	return &proxy{
		nameResolver:                  k8sconsts.NewNameResolver(config.RemoteEnvironment, config.Namespace),
		serviceDefService:             serviceDefService,
		cache:                         NewCache(config.ProxyCacheTTL),
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

	cacheEntry, err := p.getOrCreateCacheEntry(id)
	if err != nil {
		handleErrors(w, err)
		return
	}

	newRequest, cancel := p.prepareRequest(r, cacheEntry)
	defer cancel()

	err = p.addAuthentication(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	p.addRetryHandler(newRequest, id, cacheEntry)

	cacheEntry.Proxy.ServeHTTP(w, newRequest)
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

	authenticationStrategy := p.newAuthenticationStrategy(serviceApi.Credentials)

	return p.cache.Put(id, proxy, authenticationStrategy), nil
}

func (p *proxy) prepareRequest(r *http.Request, cacheEntry *CacheEntry) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.proxyTimeout)*time.Second)
	newRequest := r.WithContext(ctx)

	return newRequest, cancel
}

func (p *proxy) addAuthentication(r *http.Request, cacheEntry *CacheEntry) apperrors.AppError {
	return cacheEntry.AuthorizationStrategy.Setup(r)
}

func (p *proxy) newAuthenticationStrategy(credentials *serviceapi.Credentials) authorization.Strategy {
	authCredentials := authorization.Credentials{}

	if oauthCredentialsProvided(credentials) {
		authCredentials = authorization.Credentials{
			Oauth: &authorization.OauthCredentials{
				ClientId:          credentials.Oauth.ClientID,
				ClientSecret:      credentials.Oauth.ClientSecret,
				AuthenticationUrl: credentials.Oauth.URL,
			},
		}
	} else if basicAuthCredentialsProvided(credentials) {
		authCredentials = authorization.Credentials{
			Basic: &authorization.BasicAuthCredentials{
				UserName: credentials.Basic.Username,
				Password: credentials.Basic.Password,
			},
		}
	}

	return p.authenticationStrategyFactory.Create(authCredentials)
}

func (p *proxy) addRetryHandler(r *http.Request, id string, cacheEntry *CacheEntry) {
	cacheEntry.Proxy.ModifyResponse = p.createRequestRetrier(id, r)
}

func (p *proxy) createRequestRetrier(id string, r *http.Request) func(*http.Response) error {
	// Handle the case when credentials has been changed or OAuth token has expired
	return func(response *http.Response) error {
		retrier := newForbiddenRequestRetrier(id, r, p.proxyTimeout, p.createCacheEntry)

		return retrier.RetryIfFailedToAuthorize(response)
	}
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
	return credentials != nil && credentials.Oauth != nil
}

func basicAuthCredentialsProvided(credentials *serviceapi.Credentials) bool {
	return credentials != nil && credentials.Basic != nil
}
