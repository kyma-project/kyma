package externalapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
)

const (
	CertUrl = "https://%s/v1/applications/%s"
	SignUrl = "https://%s/v1/applications/%s/client-certs?token=%s"
	APIUrl  = "https://gateway.%s/%s/v1/"
)

type infoHandler struct {
	tokenCache     tokencache.TokenCache
	tokenGenerator tokens.TokenGenerator
	host           string
	domainName     string
	csr            csrInfo
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
	if token == "" {
		respondWithError(w, apperrors.Forbidden("Token not provided."))
		return
	}

	reName := mux.Vars(r)["appName"]

	cachedToken, found := ih.tokenCache.Get(reName)

	if !found || cachedToken != token {
		respondWithError(w, apperrors.Forbidden("Invalid token."))
		return
	}

	newToken, err := ih.tokenGenerator.NewToken(reName)
	if err != nil {
		respondWithError(w, apperrors.Internal("Failed to generate new token."))
		return
	}

	signUrl := fmt.Sprintf(SignUrl, ih.host, reName, newToken)
	certUrl := fmt.Sprintf(CertUrl, ih.host, reName)

	apiUrl := fmt.Sprintf(APIUrl, ih.domainName, reName)
	api := api{
		MetadataURL:     apiUrl + "metadata/services",
		EventsURL:       apiUrl + "events",
		CertificatesUrl: certUrl,
	}

	certInfo := makeCertInfo(ih.csr, reName)

	respondWithBody(w, 200, infoResponse{SignUrl: signUrl, Api: api, CertificateInfo: certInfo})
}

func makeCertInfo(info csrInfo, reName string) certInfo {
	subject := fmt.Sprintf("OU=%s,O=%s,L=%s,ST=%s,C=%s,CN=%s", info.OrganizationalUnit, info.Organization, info.Locality, info.Province, info.Country, reName)

	return certInfo{
		Subject:      subject,
		Extensions:   "",
		KeyAlgorithm: "rsa2048",
	}
}
