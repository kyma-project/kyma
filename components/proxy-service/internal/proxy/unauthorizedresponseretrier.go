package proxy

import (
	"context"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type retrier struct {
	id                       string
	request                  *http.Request
	retried                  bool
	timeout                  int
	updateCacheEntryFunction updateCacheEntryFunction
}

type updateCacheEntryFunction = func(string) (*CacheEntry, apperrors.AppError)

func newUnathorizedResponseRetrier(id string, request *http.Request, timeout int, updateCacheEntryFunc updateCacheEntryFunction) *retrier {
	return &retrier{id: id, request: request, retried: false, timeout: timeout, updateCacheEntryFunction: updateCacheEntryFunc}
}

func (rr *retrier) RetryIfFailedToAuthorize(r *http.Response) error {
	if rr.retried {
		return nil
	}

	rr.retried = true

	if r.StatusCode == http.StatusForbidden || r.StatusCode == http.StatusUnauthorized {
		log.Infof("Request from service with id %s failed with %d status, invalidating proxy and retrying.", rr.id, r.StatusCode)

		res, err := rr.retry()
		if err != nil {
			return err
		}

		if res != nil {
			if r.Body != nil {
				r.Body.Close()
			}
			*r = *res
		}

	}

	return nil
}

func (rr *retrier) retry() (*http.Response, error) {
	request, cancel := rr.prepareRequest()
	defer cancel()

	var err error

	cacheEntry, err := rr.updateCacheEntryFunction(rr.id)
	if err != nil {
		return nil, err
	}

	err = rr.addAuthorization(request, cacheEntry)
	if err != nil {
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

	return authorizationStrategy.AddAuthorizationHeader(r)
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
