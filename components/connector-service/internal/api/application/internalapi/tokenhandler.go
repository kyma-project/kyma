package internalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/kyma-project/kyma/components/connector-service/internal/uuid"

	"github.com/kyma-project/kyma/components/connector-service/internal/verification"

	"github.com/gorilla/mux"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const TokenURL = "https://%s/v1/applications/%s/info?token=%s"

type tokenHandler struct {
	tokenGenerator  tokens.Service
	host            string
	verificationSvc verification.Service
	uuidGenerator   uuid.Generator
}

func NewTokenHandler(varificationSvc verification.Service, tokenService tokens.Service, host string) TokenHandler {
	return &tokenHandler{
		tokenGenerator:  tokenService,
		host:            host,
		verificationSvc: varificationSvc,
	}
}

func (tg *tokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	identifier := tg.getIdentifier(r)

	tokenData, err := tg.verificationSvc.Verify(r, identifier)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	token, err := tg.tokenGenerator.CreateToken(identifier, tokenData)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	url := fmt.Sprintf(TokenURL, tg.host, identifier, token)
	res := api.TokenResponse{URL: url, Token: token}

	api.RespondWithBody(w, 201, res)
}

func (tg *tokenHandler) getIdentifier(req *http.Request) string {
	appName := mux.Vars(req)["appName"]

	if appName != "" {
		return appName
	}

	return tg.uuidGenerator.NewUUID()
}
