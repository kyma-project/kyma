package authorization

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type oauthStrategy struct {
	oauthClient            OAuthClient
	clientId               string
	clientSecret           string
	url                    string
	requestParameters      *RequestParameters
	tokenRequestSkipVerify bool
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

func (o oauthStrategy) AddAuthorization(r *http.Request, _ clientcert.SetClientCertificateFunc, skipTLSVerification bool) apperrors.AppError {
	headers, queryParameters := o.requestParameters.unpack()
	token, err := o.oauthClient.GetToken(o.clientId, o.clientSecret, o.url, headers, queryParameters, skipTLSVerification)
	if err != nil {
		zap.L().Error("failed to get token",
			zap.Error(err))
		return err
	}

	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	return nil
}

func (o oauthStrategy) Invalidate() {
	o.oauthClient.InvalidateTokenCache(o.clientId, o.clientSecret, o.url)
}
