package authorization

import (
	"crypto/tls"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/clientcert"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/oauth/tokencache"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/httpconsts"
)

type oauthWithCertStrategy struct {
	tokenCache        tokencache.TokenCache
	oauthClient       OAuthClient
	clientId          string
	certificate       []byte
	privateKey        []byte
	url               string
	requestParameters *RequestParameters
}

func newOAuthWithCertStrategy(oauthClient OAuthClient, clientId string, certificate, privateKey []byte, url string, requestParameters *RequestParameters, tokenCache tokencache.TokenCache) oauthWithCertStrategy {
	return oauthWithCertStrategy{
		tokenCache:        tokenCache,
		oauthClient:       oauthClient,
		clientId:          clientId,
		certificate:       certificate,
		privateKey:        privateKey,
		url:               url,
		requestParameters: requestParameters,
	}
}

func (o oauthWithCertStrategy) AddAuthorization(r *http.Request, _ clientcert.SetClientCertificateFunc) apperrors.AppError {
	cert, err := o.prepareCertificate()
	if err != nil {
		return apperrors.Internal("Failed to prepare certificate, %s", err.Error())
	}
	headers, queryParameters := o.requestParameters.unpack()
	token, err := o.oauthClient.GetTokenMTLS(o.clientId, o.url, cert, headers, queryParameters, o.tokenCache)
	if err != nil {
		log.Errorf("failed to get token : '%s'", err)
		return apperrors.Internal("Failed to get token: %s", err.Error())
	}

	r.Header.Set(httpconsts.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

	return nil
}

func (o oauthWithCertStrategy) Invalidate() {
	o.oauthClient.InvalidateTokenCache(o.clientId, o.url)
}

func (o oauthWithCertStrategy) prepareCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(o.certificate, o.privateKey)
}
