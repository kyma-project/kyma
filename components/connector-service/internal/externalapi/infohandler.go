package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

const (
	CsrURLFormat = "https://%s/v1/applications/%s/client-certs?token=%s"
)

type infoHandler struct {
	tokenCreator        tokens.Creator
	csr                 csrInfo
	serializerExtractor httpcontext.SerializerExtractor
	apiInfoUrlsStrategy APIUrlsGenerator
}

func NewInfoHandler(cache tokencache.TokenCache, tokenGenerator tokens.TokenGenerator, host string, domainName string, subjectValues certificates.CSRSubject) InfoHandler {
	csr := csrInfo{
		Country:            subjectValues.Country,
		Organization:       subjectValues.Organization,
		OrganizationalUnit: subjectValues.OrganizationalUnit,
		Locality:           subjectValues.Locality,
		Province:           subjectValues.Province,
	}

	return &infoHandler{tokenCache: cache, tokenGenerator: tokenGenerator, host: host, domainName: domainName, csr: csr}
}

func (ih *infoHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	serializableContext, err := ih.serializerExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	newToken, err := ih.tokenCreator.Replace(token, serializableContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	csrURL := fmt.Sprintf(CsrURLFormat, ih.host, reName, newToken)
	apiURLs := ih.apiInfoUrlsStrategy.Generate(serializableContext)

	certInfo := makeCertInfo(ih.csr, reName)

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
