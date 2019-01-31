package externalapi

import (
	"fmt"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
)

const (
	RenewCertURLFormat = "https://%s/v1/applications/certificates/renewals"
)

type ManagementInfoHandler struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
	connectorServiceHost     string
}

func NewManagementInfoHandler(connectorClientExtractor clientcontext.ConnectorClientExtractor, connectorServiceHost string) *ManagementInfoHandler {
	return &ManagementInfoHandler{
		connectorClientExtractor: connectorClientExtractor,
		connectorServiceHost:     connectorServiceHost,
	}
}

func (ih *ManagementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {
	connectorClientContext, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	urls := ih.buildURLs(connectorClientContext)

	httphelpers.RespondWithBody(w, 200, mgmtInfoReponse{URLs: urls})
}

func (ih *ManagementInfoHandler) buildURLs(connectorClientContext clientcontext.ConnectorClientContext) mgmtURLs {
	return mgmtURLs{
		RuntimeURLs:  connectorClientContext.GetRuntimeUrls(),
		RenewCertURL: fmt.Sprintf(RenewCertURLFormat, ih.connectorServiceHost),
	}
}
