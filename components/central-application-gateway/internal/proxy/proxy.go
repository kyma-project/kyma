package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/proxyconfig"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/httperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type proxy struct {
	serviceDefService            metadata.ServiceDefinitionService
	cache                        Cache
	skipVerify                   bool
	proxyTimeout                 int
	authorizationStrategyFactory authorization.StrategyFactory
	csrfTokenStrategyFactory     csrf.TokenStrategyFactory
}

// Handler serves as a Reverse Proxy
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Config stores Proxy config
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
	config Config) Handler {
	return &proxy{
		serviceDefService:            serviceDefService,
		cache:                        NewCache(config.ProxyCacheTTL),
		skipVerify:                   config.SkipVerify,
		proxyTimeout:                 config.ProxyTimeout,
		authorizationStrategyFactory: authorizationStrategyFactory,
		csrfTokenStrategyFactory:     csrfTokenStrategyFactory,
	}
}

// NewInvalidStateHandler creates handler always returning 500 response
func NewInvalidStateHandler(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleErrors(w, apperrors.Internal(message))
	})
}

// wcześniej tutaj braliśmy wartość z naglowka http
// teraz bedziemy patrzyc tylko na sciezke
func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//appName, id, path := extractAppInfo(r.URL.Path)
	// just first approach MPS
	appName, serviceName, apiName, path := extractAppInfoX(r.URL.Path)

	r.URL.Path = path // tutaj nowa scieżka

	cacheEntry, err := p.getOrCreateCacheEntry(appName, serviceName, apiName)
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

	if err := p.addModifyResponseHandler(newRequest, appName, serviceName, apiName, cacheEntry, p.createCacheEntry); err != nil {
		handleErrors(w, err)
		return
	}

	cacheEntry.Proxy.ServeHTTP(w, newRequest)
}

func (p *proxy) getOrCreateCacheEntry(appName, serviceName, apiName string) (*CacheEntry, apperrors.AppError) {
	cacheObj, found := p.cache.Get(appName, serviceName, apiName)

	if found {
		return cacheObj, nil
	}

	return p.createCacheEntry(appName, serviceName, apiName)
}

// OK wszedzie wszedzie wszedzie zmiana serviceId na serviceName
// God help us and make this name unique!!!
func (p *proxy) createCacheEntry(appName, serviceName, apiName string) (*CacheEntry, apperrors.AppError) {
	serviceAPI, err := p.serviceDefService.GetAPI(appName, serviceName, apiName)
	if err != nil {
		return nil, err
	}

	proxy, err := makeProxy(serviceAPI.TargetUrl, serviceAPI.RequestParameters, serviceName, p.skipVerify)
	if err != nil {
		return nil, err
	}

	authorizationStrategy := p.newAuthorizationStrategy(serviceAPI.Credentials)
	csrfTokenStrategy := p.newCSRFTokenStrategy(authorizationStrategy, serviceAPI.Credentials)

	return p.cache.Put(appName, serviceName, apiName, proxy, authorizationStrategy, csrfTokenStrategy), nil
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

func (p *proxy) newCSRFTokenStrategyFromCSRFConfig(authorizationStrategy authorization.Strategy, csrfConfig *proxyconfig.CSRFConfig) csrf.TokenStrategy {
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

func (p *proxy) addModifyResponseHandler(r *http.Request, appName, serviceName, apiName string, cacheEntry *CacheEntry, cacheUpdateFunc updateCacheEntryFunction) apperrors.AppError {
	// Handle the case when credentials has been changed or OAuth token has expired
	secondRequestBody, err := copyRequestBody(r)
	if err != nil {
		return err
	}

	modifyResponseFunction := func(response *http.Response) error {
		retrier := newUnauthorizedResponseRetrier(appName, serviceName, apiName, r, secondRequestBody, p.proxyTimeout, cacheUpdateFunc)
		return retrier.RetryIfFailedToAuthorize(response)
	}

	cacheEntry.Proxy.ModifyResponse = modifyResponseFunction
	return nil
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

func logAndHandleErrors(log *logrus.Entry, w http.ResponseWriter, apperr apperrors.AppError) {
	log.Errorf(apperr.Error())
	handleErrors(w, apperr)
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

func extractAppInfo(path string) (appName, serviceID, finalPath string) {
	split := strings.Split(path, "/")
	appName = split[1]
	serviceID = split[2]
	return string(appName), string(serviceID), strings.Join(split[3:], "/")
}

// {SERVICE-NAME} OS
// {BUNDLE-NAME}/{API-NAME} SKR
//
//- `{APP-NAME}/{SERVICE-NAME}/{TARGET-PATH}`
//- `{APP-NAME}/{BUNDLE-NAME}/{API-NAME}/{TARGET-PATH}`
//               service-name/ entry / path

//- `{APP-NAME}/{SERVICE-NAME}/{TARGET-PATH}`
//- `{APP-NAME}/{SERVICE-NAME}/{ENTRY-NAME}/{TARGET-PATH}`

// W OS bierzemy pierwszy entry
// W SKR by wyszukać entry musimy użyć nazwy apiName jako klucza

func extractAppInfoOS(path string) (appName, serviceName, finalPath string) {
	split := strings.Split(path, "/")

	appName = split[1]
	serviceName = split[2]
	return string(appName), string(serviceName), strings.Join(split[3:], "/")
}

func extractAppInfoX(path string) (appName, serviceName, apiName, finalPath string) {
	split := strings.Split(path, "/")

	appName = split[1]
	serviceName = split[2]
	apiName = split[3]
	return string(appName), string(serviceName), string(apiName), strings.Join(split[4:], "/")
}

// Service represents part of the remote environment, which is mapped 1 to 1 in the service-catalog to:
// - service class in V1
// - service plans in V2 (since api-packages support)

//SKR
//service     apiBundle  (apiPackage)
//  entry1      apiDefinition
//  entry2      apiDefinition
//  entry3      apiDefinition

// Kyma OS
// serviceName - wcześniej było sobie serviceID to będzie nasze API name

//service     (apiPackage)
//  entry1      apiDefinition

//service2
//  entry1      apiDefinition

// W SKR Service reprzentuje bundle a entry reprezentuje API definition
// W OS Entry jest target URL/ typ API i credentiale

// application-gateway.kyma-integration/{APP-NAME}/{SERVICE-NAME}/{TARGET-PATH}
// there will be following URL {APP-NAME}/{SERVICE-NAME}/{TARGET-PATH}
//The API contains two endpoints:
//- `{APP-NAME}/{SERVICE-NAME}/{TARGET-PATH}`
//- `{APP-NAME}/{BUNDLE-NAME}/{API-NAME}/{TARGET-PATH}`

//APP-NAME}/{SERVICE-ID}/{TARGET-PATH}

// the Application Gateway caches the OAuth tokens and CSRF tokens it obtains.
//If the service doesn't find valid tokens for the call it makes, it gets new tokens from the OAuth server and the CSRF token endpoint.
