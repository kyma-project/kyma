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
	CertificateURLFormat = "%s?token=%s"
)

type CSRInfoHandler struct {
	tokenManager             tokens.Manager
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	apiInfoURLsGenerator     APIUrlsGenerator
	certificateURL           string
	csrSubject               certificates.CSRSubject
}

func NewCSRInfoHandler(tokenManager tokens.Manager, connectorClientExtractor clientcontext.ConnectorClientExtractor, apiInfoURLsGenerator APIUrlsGenerator, certificateURL string, subjectValues certificates.CSRSubject) InfoHandler {

	return &CSRInfoHandler{
		tokenManager:             tokenManager,
		connectorClientExtractor: connectorClientExtractor,
		apiInfoURLsGenerator:     apiInfoURLsGenerator,
		certificateURL:           certificateURL,
		csrSubject:               subjectValues,
	}
}

func (ih *CSRInfoHandler) GetCSRInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	connectorClientContext, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	newToken, err := ih.tokenManager.Replace(token, connectorClientContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	csrURL := fmt.Sprintf(CertificateURLFormat, ih.certificateURL, newToken)
	apiURLs := ih.apiInfoURLsGenerator.Generate(connectorClientContext)

	certInfo := makeCertInfo(ih.csrSubject, connectorClientContext.GetCommonName())

	httphelpers.RespondWithBody(w, 200, infoResponse{CsrURL: csrURL, API: apiURLs, CertificateInfo: certInfo})
}

func makeCertInfo(csrSubject certificates.CSRSubject, commonName string) certInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", csrSubject.OrganizationalUnit, csrSubject.Organization, csrSubject.Locality, csrSubject.Province, csrSubject.Country, commonName)

	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
