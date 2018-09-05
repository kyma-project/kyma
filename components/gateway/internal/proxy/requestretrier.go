package proxy

import (
	"context"
	"net/http"
	"time"

	"fmt"
	"github.com/kyma-project/kyma/components/gateway/internal/apperrors"
	log "github.com/sirupsen/logrus"
)

type requestRetrier struct {
	id      string
	proxy   *proxy
	request *http.Request
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

		r = res
	}

	return nil
}

func (rr *requestRetrier) invalidateAndRetry() (*http.Response, error) {
	cacheObj, appError := rr.proxy.createAndCacheProxy(rr.id)
	if appError != nil {
		return nil, appError
	}

	url := fmt.Sprintf("http://%s%s", rr.request.Host, rr.request.RequestURI)

	request, err := http.NewRequest(rr.request.Method, url, rr.request.Body)
	if err != nil {
		return nil, apperrors.Internal("Failed to create proxy request: %s", err.Error())
	}

	for header, values := range rr.request.Header {
		for _, value := range values {
			request.Header.Add(header, value)
		}
	}

	request, appError = rr.proxy.handleHeaders(request, cacheObj)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.proxy.proxyTimeout)*time.Second)
	defer cancel()

	requestWithContext := request.WithContext(ctx)

	log.Info("Retrying request")

	client := &http.Client{}
	response, err := client.Do(requestWithContext)
	if err != nil {
		return nil, err
	}

	log.Infof("Result: %s", response.StatusCode)
	log.Infof("Returning response: %s", response)

	return response, nil
}
