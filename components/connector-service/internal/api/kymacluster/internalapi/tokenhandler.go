package internalapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/kyma-project/kyma/components/connector-service/internal/uuid"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURL = "https://%s/v1/clusters/%s/info?token=%s"

type tokenHandler struct {
	tokenService  tokens.ClusterService
	host          string
	uuidGenerator uuid.Generator
}

func NewTokenHandler(tokenService tokens.ClusterService, host string, uuidGenerator uuid.Generator) TokenHandler {
	return &tokenHandler{
		tokenService:  tokenService,
		host:          host,
		uuidGenerator: uuidGenerator,
	}
}

func (tg *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	identifier := mux.Vars(r)["identifier"]

	token, err := tg.tokenService.CreateClusterToken(identifier)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURL, tg.host, identifier, token)
	res := api.TokenResponse{URL: url, Token: token}

	api.RespondWithBody(w, 201, res)
}
