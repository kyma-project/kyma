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
	CertificateURLFormat   = "%s?token=%s"
	BaseEventsPathHeader   = "Base-Events-Path"
	BaseMetadataPathHeader = "Base-Metadata-Path"

	AppURLFormat     = "https://%s/v1/applications/%s"
	RuntimeURLFormat = "https://%s/v1/runtimes/%s"

	CertsEndpoint          = "certificates"
	ManagementInfoEndpoint = "management/info"
)

type CSRInfoHandler struct {
	tokenCreator             tokens.Creator
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	certificateURL           string
	getInfoURL               string
	connectorServiceHost     string
	csrSubject               certificates.CSRSubject
	urlFormat                string
}

func NewCSRInfoHandler(tokenCreator tokens.Creator, connectorClientExtractor clientcontext.ConnectorClientExtractor, certificateURL, getInfoURL, connectorServiceHost string, subjectValues certificates.CSRSubject, urlFormat string) InfoHandler {

	return &CSRInfoHandler{
		tokenCreator:             tokenCreator,
		connectorClientExtractor: connectorClientExtractor,
		certificateURL:           certificateURL,
		getInfoURL:               getInfoURL,
		connectorServiceHost:     connectorServiceHost,
		csrSubject:               subjectValues,
		urlFormat:                urlFormat,
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

	host := ih.connectorServiceHost
	infoURL := ih.getInfoURL

	if infoURL == "" {
		infoURL = fmt.Sprintf(ih.urlFormat, host, ManagementInfoEndpoint)
	}

	apiURLs := api{
		CertificatesURL: fmt.Sprintf(ih.urlFormat, host, CertsEndpoint),
		InfoURL:         infoURL,
		RuntimeURLs:     connectorClientContext.GetRuntimeUrls(),
	}

	csrURL := fmt.Sprintf(CertificateURLFormat, ih.certificateURL, newToken)

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
