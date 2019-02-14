package externalapi

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"
	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"
	"net/http"
)

const (
	RenewCertEndpoint = "certificates/renewals"
)

type managementInfoHandler struct {
	connectorClientExtractor clientcontext.ConnectorClientExtractor
}

func NewManagementInfoHandler(connectorClientExtractor clientcontext.ConnectorClientExtractor) *managementInfoHandler {
	return &managementInfoHandler{
		connectorClientExtractor: connectorClientExtractor,
	}
}

func (ih *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {
	connectorClientContext, err := ih.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	urls := ih.buildURLs(connectorClientContext)

	httphelpers.RespondWithBody(w, 200, mgmtInfoReponse{URLs: urls})
}

func (ih *managementInfoHandler) buildURLs(connectorClientContext clientcontext.ConnectorClientContext) mgmtURLs {
	return mgmtURLs{
		RuntimeURLs:  connectorClientContext.GetRuntimeUrls(),
		RenewCertURL: "",
	}
}
