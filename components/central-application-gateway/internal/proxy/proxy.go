package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/httperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/internal/metadata/model"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/proxyconfig"
)

type proxy struct {
	cache                        Cache
	skipVerify                   bool
	proxyTimeout                 int
	authorizationStrategyFactory authorization.StrategyFactory
	csrfTokenStrategyFactory     csrf.TokenStrategyFactory
	extractPathFunc              pathExtractorFunc
	apiExtractor                 APIExtractor
}

//go:generate mockery --name=APIExtractor
type APIExtractor interface {
	Get(identifier model.APIIdentifier) (*model.API, apperrors.AppError)
}

// Config stores Proxy config
type Config struct {
	SkipVerify    bool
	ProxyTimeout  int
	Application   string
	ProxyCacheTTL int
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiIdentifier, path, err := p.extractPath(r.URL.Path)
	if err != nil {
		handleErrors(w, err)
		return
	}

	r.URL.Path = path

	cacheEntry, err := p.getOrCreateCacheEntry(apiIdentifier)
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
	cacheEntry.Proxy.ServeHTTP(w, newRequest)
}

func (p *proxy) extractPath(path string) (model.APIIdentifier, string, apperrors.AppError) {
	apiIdentifier, path, err := p.extractPathFunc(path)
	if err != nil {
		return model.APIIdentifier{}, "", apperrors.Internal("failed to extract API Identifier from path")
	}

	return apiIdentifier, path, nil
}

func (p *proxy) getOrCreateCacheEntry(apiIdentifier model.APIIdentifier) (*CacheEntry, apperrors.AppError) {
	cacheObj, found := p.cache.Get(apiIdentifier.Application, apiIdentifier.Service, apiIdentifier.Entry)

	if found {
		return cacheObj, nil
	}

	return p.createCacheEntry(apiIdentifier)
}

func (p *proxy) createCacheEntry(apiIdentifier model.APIIdentifier) (*CacheEntry, apperrors.AppError) {
	serviceAPI, err := p.apiExtractor.Get(apiIdentifier)
	if err != nil {
		return nil, err
	}
	clientCertificate := clientcert.NewClientCertificate(nil)
	authorizationStrategy := p.newAuthorizationStrategy(serviceAPI.Credentials)
	csrfTokenStrategy := p.newCSRFTokenStrategy(authorizationStrategy, serviceAPI.Credentials)
	proxy, err := makeProxy(serviceAPI.TargetUrl, serviceAPI.RequestParameters, apiIdentifier, p.skipVerify, authorizationStrategy, csrfTokenStrategy, clientCertificate, p.proxyTimeout)
	if err != nil {
		return nil, err
	}

	return p.cache.Put(apiIdentifier.Application, apiIdentifier.Service, apiIdentifier.Entry, proxy, authorizationStrategy, csrfTokenStrategy, clientCertificate), nil
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
