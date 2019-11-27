package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/internal/proxy/passport"

	"github.com/kyma-project/kyma/components/application-gateway/internal/csrf"

	"github.com/kyma-project/kyma/components/application-gateway/internal/httperrors"
	"github.com/kyma-project/kyma/components/application-gateway/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
)

type proxy struct {
	nameResolver                 k8sconsts.NameResolver
	serviceDefService            metadata.ServiceDefinitionService
	cache                        Cache
	skipVerify                   bool
	proxyTimeout                 int
	authorizationStrategyFactory authorization.StrategyFactory
	csrfTokenStrategyFactory     csrf.TokenStrategyFactory
	passportAnnotater            *passport.RequestEnricher
	storageKeyName               string
}

type Config struct {
	SkipVerify              bool
	ProxyTimeout            int
	Application             string
	ProxyCacheTTL           int
	AnnotatePassportHeaders bool
	RedisURL                string
	StorageKeyName          string
}

// New creates proxy for handling user's services calls
func New(serviceDefService metadata.ServiceDefinitionService, authorizationStrategyFactory authorization.StrategyFactory, csrfTokenStrategyFactory csrf.TokenStrategyFactory, config Config) http.Handler {
	var passportEnricher *passport.RequestEnricher
	if config.AnnotatePassportHeaders {
		passportEnricher = passport.New(config.RedisURL)
	} else {
		passportEnricher = nil
	}

	return &proxy{
		nameResolver:                 k8sconsts.NewNameResolver(config.Application),
		serviceDefService:            serviceDefService,
		cache:                        NewCache(config.ProxyCacheTTL),
		skipVerify:                   config.SkipVerify,
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
		passportAnnotater:            passportEnricher,
		storageKeyName:               config.StorageKeyName,
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

	if p.passportAnnotater != nil {
		p.passportAnnotater.AnnotatePassportHeaders(newRequest, p.storageKeyName)
	}

	err = p.addAuthorization(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	if err := p.addModifyResponseHandler(newRequest, id, cacheEntry); err != nil {
		handleErrors(w, err)
		return
	}

	p.executeRequest(w, newRequest, cacheEntry)
}

func (p *proxy) extractServiceId(host string) string {
	return p.nameResolver.ExtractServiceId(host)
}

func (p *proxy) getOrCreateCacheEntry(id string) (*CacheEntry, apperrors.AppError) {
	cacheObj, found := p.cache.Get(id)

	if found {
		return cacheObj, nil
	}

	return p.createCacheEntry(id)
}

func (p *proxy) createCacheEntry(id string) (*CacheEntry, apperrors.AppError) {
	serviceApi, err := p.serviceDefService.GetAPI(id)
	if err != nil {
		return nil, err
	}

	proxy, err := makeProxy(serviceApi.TargetUrl, serviceApi.RequestParameters, id, p.skipVerify)
	if err != nil {
		return nil, err
	}

	authorizationStrategy := p.newAuthorizationStrategy(serviceApi.Credentials)
	csrfTokenStrategy := p.newCSRFTokenStrategy(authorizationStrategy, serviceApi.Credentials)

	return p.cache.Put(id, proxy, authorizationStrategy, csrfTokenStrategy), nil
}

func (p *proxy) newAuthorizationStrategy(credentials *authorization.Credentials) authorization.Strategy {
	return p.authorizationStrategyFactory.Create(credentials)
}

func (p *proxy) newCSRFTokenStrategy(authorizationStrategy authorization.Strategy, credentials *authorization.Credentials) csrf.TokenStrategy {
	csrfTokenEndpointURL := ""
	if credentials != nil {
		csrfTokenEndpointURL = credentials.CSRFTokenEndpointURL
	}
	return p.csrfTokenStrategyFactory.Create(authorizationStrategy, csrfTokenEndpointURL)
}

func (p *proxy) prepareRequest(r *http.Request, cacheEntry *CacheEntry) (*http.Request, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.proxyTimeout)*time.Second)
	newRequest := r.WithContext(ctx)

	return newRequest, cancel
}

func (p *proxy) addAuthorization(r *http.Request, cacheEntry *CacheEntry) apperrors.AppError {

	err := cacheEntry.AuthorizationStrategy.AddAuthorization(r)

	if err != nil {
		return err
	}

	return cacheEntry.CSRFTokenStrategy.AddCSRFToken(r)
}

func (p *proxy) addModifyResponseHandler(r *http.Request, id string, cacheEntry *CacheEntry) apperrors.AppError {
	modifyResponseFunction, err := p.createModifyResponseFunction(id, r)
	if err != nil {
		return err
	}

	cacheEntry.Proxy.ModifyResponse = modifyResponseFunction
	return nil
}

func (p *proxy) createModifyResponseFunction(id string, r *http.Request) (func(*http.Response) error, apperrors.AppError) {
	// Handle the case when credentials has been changed or OAuth token has expired
	secondRequestBody, err := copyRequestBody(r)
	if err != nil {
		return nil, err
	}

	modifyResponseFunction := func(response *http.Response) error {
		retrier := newUnauthorizedResponseRetrier(id, r, secondRequestBody, p.proxyTimeout, p.createCacheEntry)
		return retrier.RetryIfFailedToAuthorize(response)
	}

	return modifyResponseFunction, nil
}

func copyRequestBody(r *http.Request) (io.ReadCloser, apperrors.AppError) {
	if r.Body == nil {
		return nil, nil
	}

	bodyCopy, secondRequestBody, err := drainBody(r.Body)
	if err != nil {
		return nil, apperrors.Internal("failed to drain request body, %s", err)
	}
	r.Body = bodyCopy

	return secondRequestBody, nil
}

func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
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
