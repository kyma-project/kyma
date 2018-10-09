package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/proxy-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata"
	"github.com/kyma-project/kyma/components/proxy-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/proxycache"
	log "github.com/sirupsen/logrus"
)

type proxy struct {
	nameResolver      k8sconsts.NameResolver
	serviceDefService metadata.ServiceDefinitionService
	oauthClient       OAuthClient
	httpProxyCache    proxycache.HTTPProxyCache
	skipVerify        bool
	proxyTimeout      int
}

// New creates proxy for handling user's services calls
func New(nameResolver k8sconsts.NameResolver, serviceDefService metadata.ServiceDefinitionService,
	oauthClient OAuthClient, httpProxyCache proxycache.HTTPProxyCache, skipVerify bool, proxyTimeout int) http.Handler {
	return &proxy{
		nameResolver:      nameResolver,
		serviceDefService: serviceDefService,
		oauthClient:       oauthClient,
		httpProxyCache:    httpProxyCache,
		skipVerify:        skipVerify,
		proxyTimeout:      proxyTimeout,
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

	cacheObj, found := p.httpProxyCache.Get(id)

	var err apperrors.AppError
	if !found {
		cacheObj, err = p.createAndCacheProxy(id)
		if err != nil {
			handleErrors(w, err)
			return
		}
	}

	_, err = p.handleAuthHeaders(r, cacheObj)
	if err != nil {
		handleErrors(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.proxyTimeout)*time.Second)
	defer cancel()
	requestWithContext := r.WithContext(ctx)

	rr := newRequestRetrier(id, p, r)
	cacheObj.Proxy.ModifyResponse = rr.CheckResponse

	cacheObj.Proxy.ServeHTTP(w, requestWithContext)
}

func (p *proxy) createAndCacheProxy(id string) (*proxycache.Proxy, apperrors.AppError) {
	serviceApi, err := p.serviceDefService.GetAPI(id)
	if err != nil {
		return nil, err
	}

	proxy, err := makeProxy(serviceApi.TargetUrl, id, p.skipVerify)

	if oauthCredentialsProvided(serviceApi.Credentials) {
		return p.httpProxyCache.Add(
			id,
			serviceApi.Credentials.Oauth.URL,
			serviceApi.Credentials.Oauth.ClientID,
			serviceApi.Credentials.Oauth.ClientSecret,
			proxy,
		), nil
	}

	return p.httpProxyCache.Add(
		id,
		"",
		"",
		"",
		proxy,
	), nil
}

func (p *proxy) handleAuthHeaders(r *http.Request, cacheObj *proxycache.Proxy) (*http.Request, apperrors.AppError) {
	kymaAuthorization := handleKymaAuthorization(r)

	if !kymaAuthorization && cacheObj.OauthUrl != "" {
		err := p.addCredentials(r, cacheObj.OauthUrl, cacheObj.ClientId, cacheObj.ClientSecret)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (p *proxy) addCredentials(r *http.Request, oauthUrl, clientId, clientSecret string) apperrors.AppError {
	token, err := p.oauthClient.GetToken(clientId, clientSecret, oauthUrl)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, token)
	log.Infof("OAuth token fetched. Adding Authorization header: %s", r.Header.Get(httpconsts.HeaderAuthorization))

	return nil
}

func (p *proxy) invalidateAndHandleAuthHeaders(r *http.Request, cacheObj *proxycache.Proxy) (*http.Request, apperrors.AppError) {
	kymaAuthorization := handleKymaAuthorization(r)

	if !kymaAuthorization && cacheObj.OauthUrl != "" {
		err := p.invalidateAndAddCredentials(r, cacheObj.OauthUrl, cacheObj.ClientId, cacheObj.ClientSecret)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (p *proxy) invalidateAndAddCredentials(r *http.Request, oauthUrl, clientId, clientSecret string) apperrors.AppError {
	token, err := p.oauthClient.InvalidateAndRetry(clientId, clientSecret, oauthUrl)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, token)
	log.Infof("OAuth token fetched. Adding Authorization header: %s", r.Header.Get("Authorization"))

	return nil
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

func handleKymaAuthorization(r *http.Request) bool {
	kymaAuthorization := r.Header.Get(httpconsts.HeaderAccessToken)
	if kymaAuthorization != "" {
		r.Header.Del(httpconsts.HeaderAccessToken)
		r.Header.Set(httpconsts.HeaderAuthorization, kymaAuthorization)
		return true
	}

	return false
}
