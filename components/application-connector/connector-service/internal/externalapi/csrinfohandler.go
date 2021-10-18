package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/tokens"
)

const (
	TokenFormat   = "?token=%s"
	CertsEndpoint = "/certificates"
)

type csrInfoHandler struct {
	tokenManager             tokens.Creator
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	getInfoURL               string
	baseURL                  string
	csrSubject               certificates.CSRSubject
}

func NewCSRInfoHandler(tokenManager tokens.Creator, connectorClientExtractor clientcontext.ConnectorClientExtractor, getInfoURL string, baseURL string) CSRInfoHandler {

	return &csrInfoHandler{
		tokenManager:             tokenManager,
		connectorClientExtractor: connectorClientExtractor,
		getInfoURL:               getInfoURL,
		baseURL:                  baseURL,
	}
}

func (ih *csrInfoHandler) GetCSRInfo(w http.ResponseWriter, r *http.Request) {
	clientContextService, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	newToken, err := ih.tokenManager.Save(clientContextService.ClientContext())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	apiURLs := ih.makeApiURLs(clientContextService)

	csrURL := ih.makeCSRURLs(newToken)

	certInfo := makeCertInfo(clientContextService.GetSubject().ToString())

	httphelpers.RespondWithBody(w, http.StatusOK, csrInfoResponse{CsrURL: csrURL, API: apiURLs, CertificateInfo: certInfo})
}

func (ih *csrInfoHandler) makeCSRURLs(newToken string) string {
	csrURL := ih.baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func (ih *csrInfoHandler) makeApiURLs(clientContextService clientcontext.ClientCertContextService) api {
	return api{
		CertificatesURL: ih.baseURL + CertsEndpoint,
		InfoURL:         ih.getInfoURL,
		RuntimeURLs:     clientContextService.GetRuntimeUrls(),
	}
}

func makeCertInfo(subject string) certInfo {
	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
