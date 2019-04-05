package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

const (
	RenewCertURLFormat      = "%s/certificates/renewals"
	RevocationCertURLFormat = "%s/certificates/revocations"
)

type managementInfoHandler struct {
	connectorClientExtractor    clientcontext.ConnectorClientExtractor
	certificateProtectedBaseURL string
	csrSubject                  certificates.CSRSubject
}

func NewManagementInfoHandler(connectorClientExtractor clientcontext.ConnectorClientExtractor, certProtectedBaseURL string, subjectValues certificates.CSRSubject) *managementInfoHandler {
	return &managementInfoHandler{
		connectorClientExtractor:    connectorClientExtractor,
		certificateProtectedBaseURL: certProtectedBaseURL,
		csrSubject:                  subjectValues,
	}
}

func (ih *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {
	clientContextService, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	urls := ih.buildURLs(clientContextService)

	certInfo := makeCertInfo(ih.csrSubject, clientContextService.GetCommonName())

	httphelpers.RespondWithBody(w, http.StatusOK, mgmtInfoReponse{URLs: urls, ClientIdentity: clientContextService, CertificateInfo: certInfo})
}

func (ih *managementInfoHandler) buildURLs(clientContextService clientcontext.ClientContextService) mgmtURLs {
	return mgmtURLs{
		RuntimeURLs:       clientContextService.GetRuntimeUrls(),
		RenewCertURL:      fmt.Sprintf(RenewCertURLFormat, ih.certificateProtectedBaseURL),
		RevocationCertURL: fmt.Sprintf(RevocationCertURLFormat, ih.certificateProtectedBaseURL),
	}
}
