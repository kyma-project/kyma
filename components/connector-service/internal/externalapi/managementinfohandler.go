package externalapi

import (
	"fmt"
	"net/http"

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
}

func NewManagementInfoHandler(connectorClientExtractor clientcontext.ConnectorClientExtractor, certProtectedBaseURL string) *managementInfoHandler {
	return &managementInfoHandler{
		connectorClientExtractor:    connectorClientExtractor,
		certificateProtectedBaseURL: certProtectedBaseURL,
	}
}

func (ih *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {
	clientContextService, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	urls := ih.buildURLs(clientContextService)

	certInfo := makeCertInfo(clientContextService.GetSubject().ToString())

	httphelpers.RespondWithBody(w, http.StatusOK, mgmtInfoReponse{URLs: urls, ClientIdentity: clientContextService.ClientContext(), CertificateInfo: certInfo})
}

func (ih *managementInfoHandler) buildURLs(clientContextService clientcontext.ClientCertContextService) mgmtURLs {
	return mgmtURLs{
		RuntimeURLs:   clientContextService.GetRuntimeUrls(),
		RenewCertURL:  fmt.Sprintf(RenewCertURLFormat, ih.certificateProtectedBaseURL),
		RevokeCertURL: fmt.Sprintf(RevocationCertURLFormat, ih.certificateProtectedBaseURL),
	}
}
