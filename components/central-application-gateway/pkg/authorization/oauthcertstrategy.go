package authorization

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
)

type oauthWithCertStrategy struct {
	oauthClient            OAuthClient
	clientId               string
	clientSecret           string
	certificate            []byte
	privateKey             []byte
	url                    string
	requestParameters      *RequestParameters
	tokenRequestSkipVerify bool
}

func newOAuthWithCertStrategy(oauthClient OAuthClient, clientId string, clientSecret string, certificate, privateKey []byte, url string, requestParameters *RequestParameters) oauthWithCertStrategy {
	return oauthWithCertStrategy{
		oauthClient:       oauthClient,
		clientId:          clientId,
		clientSecret:      clientSecret,
		certificate:       certificate,
		privateKey:        privateKey,
		url:               url,
		requestParameters: requestParameters,
	}
}

func (o oauthWithCertStrategy) AddAuthorization(r *http.Request, _ clientcert.SetClientCertificateFunc, skipTLSVerification bool) apperrors.AppError {
	log.Infof("Passing skipTLSVerification=%v to GetTokenMTLS", skipTLSVerification)
	headers, queryParameters := o.requestParameters.unpack()
	token, err := o.oauthClient.GetTokenMTLS(o.clientId, o.url, o.certificate, o.privateKey, headers, queryParameters, skipTLSVerification)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return apperrors.Internal("Failed to get token: %s", err.Error())
	}

	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	return nil
}

func (o oauthWithCertStrategy) Invalidate() {
	o.oauthClient.InvalidateTokenCacheMTLS(o.clientId, o.url, o.certificate, o.privateKey)
}
