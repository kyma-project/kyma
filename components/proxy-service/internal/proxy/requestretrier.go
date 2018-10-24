package proxy

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type requestRetrier struct {
	id      string
	proxy   *proxy
	request *http.Request
	host    string
	retried bool
}

func newRequestRetrier(id string, proxy *proxy, request *http.Request) *requestRetrier {
	return &requestRetrier{id: id, proxy: proxy, request: request, retried: false}
}

func (rr *requestRetrier) CheckResponse(r *http.Response) error {
	if rr.retried {
		return nil
	}

	rr.retried = true

	if r.StatusCode == 403 {
		log.Infof("Request from service with id %s failed with 403 status, invalidating proxy and retrying.", rr.id)

		res, err := rr.invalidateAndRetry()
		if err != nil {
			return err
		}
		if res != nil {
			r = res
		}
	}

	return nil
}

func (rr *requestRetrier) invalidateAndRetry() (*http.Response, error) {
	cacheObj, appError := rr.proxy.createAndCacheProxy(rr.id)
	if appError != nil {
		return nil, appError
	}

	if cacheObj.OauthUrl == "" {
		return nil, nil
	}

	rr.request.RequestURI = ""

	_, appError = rr.proxy.invalidateAndHandleAuthHeaders(rr.request, cacheObj)
	if appError != nil {
		return nil, appError
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.proxy.proxyTimeout)*time.Second)
	defer cancel()

	requestWithContext := rr.request.WithContext(ctx)
	cacheObj.Proxy.Director(requestWithContext)

	client := &http.Client{
		Transport: cacheObj.Proxy.Transport,
	}

	response, err := client.Do(requestWithContext)
	if err != nil {
		return nil, err
	}

	return response, nil
}
