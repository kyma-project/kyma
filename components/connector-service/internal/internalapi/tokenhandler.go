package internalapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURL = "https://%s/v1/applications/%s/info?token=%s"

type tokenHandler struct {
	tokenGenerator tokens.TokenGenerator
	host           string
}

func NewTokenHandler(tokenGenerator tokens.TokenGenerator, host string) TokenHandler {
	return &tokenHandler{tokenGenerator: tokenGenerator, host: host}
}

func (tg *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	reName := mux.Vars(r)["appName"]
	token, err := tg.tokenGenerator.NewToken(reName)
	if err != nil {
		respondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURL, tg.host, reName, token)
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
