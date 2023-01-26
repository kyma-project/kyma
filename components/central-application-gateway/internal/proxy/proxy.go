package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	proxyTimeout                 int
	authorizationStrategyFactory authorization.StrategyFactory
	csrfTokenStrategyFactory     csrf.TokenStrategyFactory
	extractPathFunc              pathExtractorFunc
	extractGatewayFunc           gatewayURLExtractorFunc
	apiExtractor                 APIExtractor
}

//go:generate mockery --name=APIExtractor
type APIExtractor interface {
	Get(identifier model.APIIdentifier) (*model.API, apperrors.AppError)
}

// Config stores Proxy config
type Config struct {
	ProxyTimeout  int
	Application   string
	ProxyCacheTTL int
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	apiIdentifier, path, gwURL, err := p.extractPath(r.URL)
	if err != nil {
		handleErrors(w, err)
		return
	}

	serviceAPI, err := p.apiExtractor.Get(apiIdentifier)
	if err != nil {
		handleErrors(w, err)
		return
	}

	r.URL.Path = path.Path
	if !serviceAPI.EncodeUrl {
		r.URL.RawPath = path.RawPath
	}

	cacheEntry, err := p.getOrCreateCacheEntry(apiIdentifier, *serviceAPI)
	if err != nil {
		handleErrors(w, err)
		return
	}

	newRequest, cancel := p.setRequestTimeout(r)
	defer cancel()

	err = p.addAuthorization(newRequest, cacheEntry, serviceAPI.SkipVerify)
	if err != nil {
		handleErrors(w, err)
		return
	}

	cacheEntry.Proxy.ModifyResponse = responseModifier(gwURL, serviceAPI.TargetUrl, urlRewriter)
	cacheEntry.Proxy.ServeHTTP(w, newRequest)
}

func (p *proxy) extractPath(u *url.URL) (model.APIIdentifier, *url.URL, *url.URL, apperrors.AppError) {
	apiIdentifier, path, gwURL, err := p.extractPathFunc(u)
	if err != nil {
		return model.APIIdentifier{}, nil, nil, apperrors.WrongInput("failed to extract API Identifier from path")
	}

	return apiIdentifier, path, gwURL, nil
}

func (p *proxy) getOrCreateCacheEntry(apiIdentifier model.APIIdentifier, serviceAPI model.API) (*CacheEntry, apperrors.AppError) {
	cacheObj, found := p.cache.Get(apiIdentifier.Application, apiIdentifier.Service, apiIdentifier.Entry)

	if found {
		return cacheObj, nil
	}

	return p.createCacheEntry(apiIdentifier, serviceAPI)
}

func (p *proxy) createCacheEntry(apiIdentifier model.APIIdentifier, serviceAPI model.API) (*CacheEntry, apperrors.AppError) {
	clientCertificate := clientcert.NewClientCertificate(nil)
	authorizationStrategy := p.newAuthorizationStrategy(serviceAPI.Credentials)
	csrfTokenStrategy := p.newCSRFTokenStrategy(authorizationStrategy, serviceAPI.Credentials)
	proxy, err := makeProxy(serviceAPI.TargetUrl, serviceAPI.RequestParameters, apiIdentifier.Service, serviceAPI.SkipVerify, authorizationStrategy, csrfTokenStrategy, clientCertificate, p.proxyTimeout)
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

func (p *proxy) addAuthorization(r *http.Request, cacheEntry *CacheEntry, skipTLSVerify bool) apperrors.AppError {

	err := cacheEntry.AuthorizationStrategy.AddAuthorization(r, skipTLSVerify)

	if err != nil {
		return err
	}

	return cacheEntry.CSRFTokenStrategy.AddCSRFToken(r, skipTLSVerify)
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
