package internalapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
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
		respondWithError(w, err)
		return
	}

	token, err := th.tokenService.Save(tokenParams)
	if err != nil {
		respondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURLFormat, th.csrInfoURL, token)
	res := tokenResponse{URL: url, Token: token}

	respondWithBody(w, 201, res)
}

func respondWithError(w http.ResponseWriter, apperr apperrors.AppError) {
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr)

	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

type tokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}
