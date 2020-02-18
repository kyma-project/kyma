package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	proxyPkg "github.com/kyma-project/kyma/components/application-gateway/pkg/proxy"

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

	configRepository proxyPkg.TargetConfigProvider
}

type ProxyHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeHTTPNamespaced(w http.ResponseWriter, r *http.Request)
}

type Config struct {
	SkipVerify    bool
	ProxyTimeout  int
	Application   string
	ProxyCacheTTL int
}

// New creates proxy for handling user's services calls
func New(
	serviceDefService metadata.ServiceDefinitionService,
	authorizationStrategyFactory authorization.StrategyFactory,
	csrfTokenStrategyFactory csrf.TokenStrategyFactory,
	config Config,
	configRepository proxyPkg.TargetConfigProvider) ProxyHandler {
	return &proxy{
		nameResolver:                 k8sconsts.NewNameResolver(config.Application),
		serviceDefService:            serviceDefService,
		cache:                        NewCache(config.ProxyCacheTTL),
		skipVerify:                   config.SkipVerify,
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
		configRepository:             configRepository,
	}
}

// NewInvalidStateHandler creates handler always returning 500 response
func NewInvalidStateHandler(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleErrors(w, apperrors.Internal(message))
	})
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := p.nameResolver.ExtractServiceId(r.Host)

	cacheEntry, err := p.getOrCreateCacheEntry(id)
	if err != nil {
		handleErrors(w, err)
		return
	}

	newRequest, cancel := p.setRequestTimeout(r)
	defer cancel()

	err = p.addAuthorization(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	if err := p.addModifyResponseHandler(newRequest, id, cacheEntry); err != nil {
		handleErrors(w, err)
		return
	}

	cacheEntry.Proxy.ServeHTTP(w, newRequest)
}

// ServeHTTPNamespaced proxies requests using data from secrets
// serviceId is composed of secret name and API name in format: {SECRET_NAME}-{API_NAME}
func (p *proxy) ServeHTTPNamespaced(w http.ResponseWriter, r *http.Request) {
	secretName, found := mux.Vars(r)["secret"]
	if !found {
		handleErrors(w, apperrors.WrongInput("secret name not specified"))
		return
	}

	apiName, found := mux.Vars(r)["apiName"]
	if !found {
		handleErrors(w, apperrors.WrongInput("API name not specified"))
		return
	}

	serviceId := fmt.Sprintf("%s-%s", secretName, apiName)

	cacheEntry, found := p.cache.Get(serviceId)
	if !found {
		proxyConfig, err := p.configRepository.GetDestinationConfig(secretName, apiName)
		if err != nil {
			handleErrors(w, err)
			return
		}

		cacheEntry, err = p.cacheEntryFromProxyConfig(serviceId, proxyConfig)
		if err != nil {
			handleErrors(w, err)
			return
		}
	}

	newRequest, cancel := p.setRequestTimeout(r)
	defer cancel()

	err := p.addAuthorization(newRequest, cacheEntry)
	if err != nil {
		handleErrors(w, err)
		return
	}

	if err := p.addModifyResponseHandler(newRequest, serviceId, cacheEntry); err != nil {
		handleErrors(w, err)
		return
	}

	cacheEntry.Proxy.ServeHTTP(w, r)
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

func (p *proxy) cacheEntryFromProxyConfig(id string, config proxyPkg.ProxyDestinationConfig) (*CacheEntry, apperrors.AppError) {
	proxy, err := makeProxy(config.Destination.URL, config.Destination.RequestParameters, id, p.skipVerify)
	if err != nil {
		return nil, err
	}

	credentials := config.Credentials.ToCredentials()

	authorizationStrategy := p.newAuthorizationStrategy(credentials)
	csrfTokenStrategy := p.newCSRFTokenStrategyFromCSRFConfig(authorizationStrategy, config.Destination.CSRFConfig)

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

func (p *proxy) newCSRFTokenStrategyFromCSRFConfig(authorizationStrategy authorization.Strategy, csrfConfig *csrf.CSRFConfig) csrf.TokenStrategy {
	csrfTokenEndpointURL := ""
	if csrfConfig != nil {
		csrfTokenEndpointURL = csrfConfig.TokenURL
	}
	return p.csrfTokenStrategyFactory.Create(authorizationStrategy, csrfTokenEndpointURL)
}

func (p *proxy) setRequestTimeout(r *http.Request) (*http.Request, context.CancelFunc) {
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

func handleErrors(w http.ResponseWriter, apperr apperrors.AppError) {
	code, body := httperrors.AppErrorToResponse(apperr)
	respondWithBody(w, code, body)
}

func respondWithBody(w http.ResponseWriter, code int, body httperrors.ErrorResponse) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)

	w.WriteHeader(code)

	json.NewEncoder(w).Encode(body)
}
