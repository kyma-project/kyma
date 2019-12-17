package authorization

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/httpconsts"
	log "github.com/sirupsen/logrus"
)

type oauthStrategy struct {
	oauthClient       OAuthClient
	clientId          string
	clientSecret      string
	url               string
	requestParameters *RequestParameters
}

func newOAuthStrategy(oauthClient OAuthClient, clientId, clientSecret, url string, requestParameters *RequestParameters) oauthStrategy {
	return oauthStrategy{
		oauthClient:       oauthClient,
		clientId:          clientId,
		clientSecret:      clientSecret,
		url:               url,
		requestParameters: requestParameters,
	}
}

func (o oauthStrategy) AddAuthorization(r *http.Request, _ TransportSetter) apperrors.AppError {
	headers, queryParameters := o.requestParameters.unpack()
	token, err := o.oauthClient.GetToken(o.clientId, o.clientSecret, o.url, headers, queryParameters)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	return nil
}

func (o oauthStrategy) Invalidate() {
	o.oauthClient.InvalidateTokenCache(o.clientId)
}
