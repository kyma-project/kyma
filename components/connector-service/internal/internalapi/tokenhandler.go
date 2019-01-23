package internalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/httpcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/httphelpers"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURLFormat = "https://%s?token=%s"

type tokenHandler struct {
	tokenCreator        tokens.Creator
	csrInfoURL          string
	serializerExtractor httpcontext.SerializerExtractor
}

func NewTokenHandler(tokenService tokens.Creator, csrInfoURL string, serializerExtractor httpcontext.SerializerExtractor) TokenHandler {
	return &tokenHandler{tokenCreator: tokenService, csrInfoURL: csrInfoURL, serializerExtractor: serializerExtractor}
}

func (th *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	tokenParams, err := th.serializerExtractor(r.Context())
	if err != nil {
		httphelpers.RespondWithError(w, err)
		return
	}

	token, err := th.tokenCreator.Save(tokenParams)
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
