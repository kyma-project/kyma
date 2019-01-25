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
	CsrURLFormat = "https://%s/v1/applications/certificates?token=%s"
)

type CSRInfoHandler struct {
	tokenCreator             tokens.Creator
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	apiInfoURLsGenerator     APIUrlsGenerator
	host                     string
	csrSubject               certificates.CSRSubject
}

func NewCSRInfoHandler(tokenCreator tokens.Creator, connectorClientExtractor clientcontext.ConnectorClientExtractor, apiInfoURLsGenerator APIUrlsGenerator, host string, subjectValues certificates.CSRSubject) InfoHandler {

	return &CSRInfoHandler{
		tokenCreator:             tokenCreator,
		connectorClientExtractor: connectorClientExtractor,
		apiInfoURLsGenerator:     apiInfoURLsGenerator,
		host:                     host,
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

	newToken, err := ih.tokenCreator.Replace(token, connectorClientContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	csrURL := fmt.Sprintf(CsrURLFormat, ih.host, newToken)
	apiURLs := ih.apiInfoURLsGenerator.Generate(connectorClientContext)

	certInfo := makeCertInfo(ih.csrSubject, connectorClientContext.GetCommonName())

	httphelpers.RespondWithBody(w, 200, infoResponse{CsrURL: csrURL, API: apiURLs, CertificateInfo: certInfo})
}

func makeCertInfo(csrSubject certificates.CSRSubject, reName string) certInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", csrSubject.OrganizationalUnit, csrSubject.Organization, csrSubject.Locality, csrSubject.Province, csrSubject.Country, reName)

	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
