package proxy

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
)

type retrier struct {
	id                       string
	appName                  string
	request                  *http.Request
	requestBodyCopy          io.ReadCloser
	retried                  bool
	timeout                  int
	updateCacheEntryFunction updateCacheEntryFunction
}

type updateCacheEntryFunction = func(string, string) (*CacheEntry, apperrors.AppError)

func newUnauthorizedResponseRetrier(appName, id string, request *http.Request, requestBodyCopy io.ReadCloser, timeout int, updateCacheEntryFunc updateCacheEntryFunction) *retrier {
	return &retrier{appName: appName, id: id, request: request, requestBodyCopy: requestBodyCopy, retried: false, timeout: timeout, updateCacheEntryFunction: updateCacheEntryFunc}
}

func (rr *retrier) RetryIfFailedToAuthorize(r *http.Response) error {
	if rr.retried {
		return nil
	}

	rr.retried = true

	if r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusUnauthorized {
		log.Infof("Request from service with id %s failed with %d status, invalidating proxy and retrying.", rr.id, r.StatusCode)

		retryRes, err := rr.retry()
		if err != nil {
			return err
		}

		if retryRes != nil {
			if r.Body != nil {
				r.Body.Close()
			}
			*r = *retryRes
		}
	}

	return nil
}

func (rr *retrier) retry() (*http.Response, error) {
	request, cancel := rr.prepareRequest()
	defer cancel()

	request.Body = rr.requestBodyCopy

	cacheEntry, err := rr.updateCacheEntryFunction(rr.appName, rr.id)
	if err != nil {
		return nil, err
	}

	if err := rr.addAuthorization(request, cacheEntry); err != nil {
		return nil, err
	}

	return rr.performRequest(request, cacheEntry)
}

func (rr *retrier) prepareRequest() (*http.Request, context.CancelFunc) {
	rr.request.RequestURI = ""
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.timeout)*time.Second)

	return rr.request.WithContext(ctx), cancel
}

func (rr *retrier) addAuthorization(r *http.Request, cacheEntry *CacheEntry) error {
	authorizationStrategy := cacheEntry.AuthorizationStrategy
	authorizationStrategy.Invalidate()
	err := cacheEntry.AuthorizationStrategy.AddAuthorization(r)
	if err != nil {
		return err
	}

	csrfTokenStrategy := cacheEntry.CSRFTokenStrategy
	csrfTokenStrategy.Invalidate()
	return csrfTokenStrategy.AddCSRFToken(r)
}

func (rr *retrier) performRequest(r *http.Request, cacheEntry *CacheEntry) (*http.Response, error) {
	reverseProxy := cacheEntry.Proxy
	reverseProxy.Director(r)

	client := &http.Client{
		Transport: reverseProxy.Transport,
	}

	res, err := client.Do(r)

	return res, err
}
