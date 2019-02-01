package internalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURLFormat = "%s?token=%s"

type tokenHandler struct {
	tokenManager             tokens.Manager
	csrInfoURL               string
	connectorClientExtractor clientcontext.ConnectorClientExtractor
}

func NewTokenHandler(tokenManager tokens.Manager, csrInfoURL string, connectorClientExtractor clientcontext.ConnectorClientExtractor) TokenHandler {
	return &tokenHandler{tokenManager: tokenManager, csrInfoURL: csrInfoURL, connectorClientExtractor: connectorClientExtractor}
}

func (th *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	connectorClientContext, err := th.connectorClientExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	token, err := th.tokenManager.Save(connectorClientContext)
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURLFormat, th.csrInfoURL, token)
	res := tokenResponse{URL: url, Token: token}

	httphelpers.RespondWithBody(w, 201, res)
}

type tokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
