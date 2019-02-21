package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
)

const (
	RenewCertURLFormat = "%s/certificates/renewals"
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
	contextServiceProvider, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	urls := ih.buildURLs(contextServiceProvider)

	httphelpers.RespondWithBody(w, http.StatusOK, mgmtInfoReponse{URLs: urls})
}

func (ih *managementInfoHandler) buildURLs(contextServiceProvider clientcontext.ClientContextService) mgmtURLs {
	return mgmtURLs{
		RuntimeURLs:  contextServiceProvider.GetRuntimeUrls(),
		RenewCertURL: fmt.Sprintf(RenewCertURLFormat, ih.certificateProtectedBaseURL),
	}
}
