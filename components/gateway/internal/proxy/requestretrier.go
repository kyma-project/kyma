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
}

func newRequestRetrier(id string, proxy *proxy, request *http.Request) *requestRetrier {
	return &requestRetrier{id: id, proxy: proxy, request: request}
}

func (rr *requestRetrier) CheckResponse(r *http.Response) error {
	if r.StatusCode == 403 {
		log.Infof("Request from service with id %s failed with 403 status, invalidating proxy and retrying.", rr.id)
		res, err := rr.invalidateAndRetry()
		if err != nil {
			return err
		}

		r = res
	}

	return nil
}

func (rr *requestRetrier) invalidateAndRetry() (*http.Response, error) {
	cacheObj, appError := rr.proxy.createAndCacheProxy(rr.id)
	if appError != nil {
		return nil, appError
	}

	request, appError := rr.proxy.handleHeaders(rr.request, cacheObj)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.proxy.proxyTimeout)*time.Second)
	defer cancel()

	requestWithContext := request.WithContext(ctx)

	client := http.Client{}

	response, err := client.Do(requestWithContext)
	if err != nil {
		return nil, err
	}

	return response, nil
}
