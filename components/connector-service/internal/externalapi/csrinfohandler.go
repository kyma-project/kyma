package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const (
	CsrURLFormat = "https://%s/v1/applications/certificates?token=%s"
)

type CSRInfoHandler struct {
	tokenCreator         tokens.Creator
	serializerExtractor  httpcontext.ConnectorClientExtractor
	apiInfoURLsGenerator APIUrlsGenerator
	host                 string
	csr                  csrInfo
}

func NewCSRInfoHandler(tokenCreator tokens.Creator, serializerExtractor httpcontext.ConnectorClientExtractor, apiInfoURLsGeneretor APIUrlsGenerator, host string, subjectValues certificates.CSRSubject) InfoHandler {
	csr := csrInfo{
		Country:            subjectValues.Country,
		Organization:       subjectValues.Organization,
		OrganizationalUnit: subjectValues.OrganizationalUnit,
		Locality:           subjectValues.Locality,
		Province:           subjectValues.Province,
	}

	return &CSRInfoHandler{
		tokenCreator:         tokenCreator,
		serializerExtractor:  serializerExtractor,
		apiInfoURLsGenerator: apiInfoURLsGeneretor,
		host:                 host,
		csr:                  csr,
	}
}

func (ih *CSRInfoHandler) GetCSRInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	kymaContext, err := ih.serializerExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	newToken, err := ih.tokenCreator.Replace(token, kymaContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	csrURL := fmt.Sprintf(CsrURLFormat, ih.host, newToken)
	apiURLs := ih.apiInfoURLsGenerator.Generate(kymaContext)

	certInfo := makeCertInfo(ih.csr, kymaContext.GetCommonName())

	httphelpers.RespondWithBody(w, 200, infoResponse{CsrURL: csrURL, API: apiURLs, CertificateInfo: certInfo})
}

func makeCertInfo(info csrInfo, reName string) certInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", info.OrganizationalUnit, info.Organization, info.Locality, info.Province, info.Country, reName)

	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
