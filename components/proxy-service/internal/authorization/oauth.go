package authorization

import (
	"net/http"
	"net/http/httputil"
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	log "github.com/sirupsen/logrus"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
)

type AuthorizationStrategy interface {
	Setup()
}

type OAuthClient interface {
	GetToken(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
	InvalidateAndRetry(clientID string, clientSecret string, authURL string) (string, apperrors.AppError)
}

type oauthStrategy struct {
	oauthClient OAuthClient
	clientId     string
	clientSecret string
	url          string
}

func (o oauthStrategy) Setup(proxy *httputil.ReverseProxy, r *http.Request) apperrors.AppError {
	token, err := o.oauthClient.GetToken(o.clientId, o.clientSecret, o.url)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, token)
	log.Infof("OAuth token fetched. Adding Authorization header: %s", r.Header.Get(httpconsts.HeaderAuthorization))

	return nil
}


type oauthRetryStrategy struct {
	oauthStrategy
	retried bool
	id string
}

func (rr *oauthRetryStrategy) Do(r *http.Response) error {
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

func (rr *oauthRetryStrategy) invalidateAndRetry() (*http.Response, error) {

	
	//_, appError = rr.proxy.invalidateAndHandleAuthHeaders(rr.request, cacheObj)
	//if appError != nil {
	//	return nil, appError
	//}
	//
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.proxy.proxyTimeout)*time.Second)
	//defer cancel()
	//
	//requestWithContext := rr.request.WithContext(ctx)
	//cacheObj.Proxy.Director(requestWithContext)
	//
	//client := &http.Client{
	//	Transport: cacheObj.Proxy.Transport,
	//}
	//
	//response, err := client.Do(requestWithContext)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return response, nil
	return nil, nil
}



type RetryStrategy interface {
	Do(r *http.Response) error
}
