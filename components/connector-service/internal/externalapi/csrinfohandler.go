package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const (
	TokenFormat   = "?token=%s"
	CertsEndpoint = "certificates"
)

type CSRInfoHandler struct {
	tokenManager             tokens.Creator
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	getInfoURL               string
	baseURL                  string
	csrSubject               certificates.CSRSubject
}

func NewCSRInfoHandler(tokenManager tokens.Creator, connectorClientExtractor clientcontext.ConnectorClientExtractor, getInfoURL string, subjectValues certificates.CSRSubject, baseURL string) CSRGetInfoHandler {

	return &CSRInfoHandler{
		tokenManager:             tokenManager,
		connectorClientExtractor: connectorClientExtractor,
		getInfoURL:               getInfoURL,
		baseURL:                  baseURL,
		csrSubject:               subjectValues,
	}
}

func (ih *CSRInfoHandler) GetCSRInfo(w http.ResponseWriter, r *http.Request) {
	connectorClientContext, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	newToken, err := ih.tokenManager.Save(connectorClientContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	apiURLs := ih.makeApiURLs(connectorClientContext)

	csrURL := ih.makeCSRURLs(newToken)

	certInfo := makeCertInfo(ih.csrSubject, connectorClientContext.GetCommonName())

	httphelpers.RespondWithBody(w, 200, csrInfoResponse{CsrURL: csrURL, API: apiURLs, CertificateInfo: certInfo})
}

func (ih *CSRInfoHandler) makeCSRURLs(newToken string) string {
	csrURL := ih.baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func (ih *CSRInfoHandler) makeApiURLs(connectorClientContext clientcontext.ConnectorClientContext) api {
	infoURL := ih.getInfoURL
	return api{
		CertificatesURL: ih.baseURL + CertsEndpoint,
		InfoURL:         infoURL,
		RuntimeURLs:     connectorClientContext.GetRuntimeUrls(),
	}
}

func makeCertInfo(csrSubject certificates.CSRSubject, commonName string) certInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", csrSubject.OrganizationalUnit, csrSubject.Organization, csrSubject.Locality, csrSubject.Province, csrSubject.Country, commonName)

	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
