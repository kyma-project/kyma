package authorization

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/tokencache"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type oauthStrategy struct {
	tokenCache        tokencache.TokenCache
	oauthClient       OAuthClient
	clientId          string
	clientSecret      string
	url               string
	requestParameters *RequestParameters
}

func newOAuthStrategy(oauthClient OAuthClient, clientId, clientSecret, url string, requestParameters *RequestParameters, tokenCache tokencache.TokenCache) oauthStrategy {
	return oauthStrategy{
		tokenCache:        tokenCache,
		oauthClient:       oauthClient,
		clientId:          clientId,
		clientSecret:      clientSecret,
		url:               url,
		requestParameters: requestParameters,
	}
}

func (o oauthStrategy) AddAuthorization(r *http.Request, _ clientcert.SetClientCertificateFunc) apperrors.AppError {
	headers, queryParameters := o.requestParameters.unpack()
	token, err := o.oauthClient.GetToken(o.clientId, o.clientSecret, o.url, headers, queryParameters, o.tokenCache)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	return nil
}

func (o oauthStrategy) Invalidate() {
	o.oauthClient.InvalidateTokenCache(o.clientId, o.url)
}
