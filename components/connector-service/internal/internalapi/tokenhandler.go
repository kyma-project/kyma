package internalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURLFormat = "https://%s?token=%s"

type tokenHandler struct {
	tokenService      tokens.Service
	csrInfoURL        string
	tokenParamsParser tokens.TokenParamsParser
}

func NewTokenHandler(tokenService tokens.Service, csrInfoURL string, tokenParamsParser tokens.TokenParamsParser) TokenHandler {
	return &tokenHandler{tokenService: tokenService, csrInfoURL: csrInfoURL, tokenParamsParser: tokenParamsParser}
}

func (th *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	tokenParams, err := th.tokenParamsParser(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	token, err := th.tokenService.Save(tokenParams)
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
