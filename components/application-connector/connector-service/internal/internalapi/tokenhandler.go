package internalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/tokens"
)

const TokenURLFormat = "%s?token=%s"

type tokenHandler struct {
	tokenManager             tokens.Creator
	csrInfoURL               string
	connectorClientExtractor clientcontext.ConnectorClientExtractor
}

func NewTokenHandler(tokenManager tokens.Creator, csrInfoURL string, connectorClientExtractor clientcontext.ConnectorClientExtractor) TokenHandler {
	return &tokenHandler{tokenManager: tokenManager, csrInfoURL: csrInfoURL, connectorClientExtractor: connectorClientExtractor}
}

func (th *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	clientContextService, err := th.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithErrorAndLog(w, err)
		return
	}

	logger := clientContextService.GetLogger()

	logger.Info("Generating token")
	token, err := th.tokenManager.Save(clientContextService.ClientContext())
	if err != nil {
		logger.Error(err)
		httphelpers.RespondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURLFormat, th.csrInfoURL, token)
	res := tokenResponse{URL: url, Token: token}

	httphelpers.RespondWithBody(w, http.StatusCreated, res)
}

type tokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
