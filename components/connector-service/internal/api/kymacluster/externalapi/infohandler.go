package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/gorilla/mux"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const (
	SignUrl = "https://%s/v1/clusters/%s/client-certs?token=%s"
)

type infoHandler struct {
	tokenService tokens.ClusterService
	host         string
	csrSubject   certificates.CSRSubject
}

func NewInfoHandler(tokenService tokens.ClusterService, host string, csrSubject certificates.CSRSubject) InfoHandler {
	return &infoHandler{
		tokenService: tokenService,
		host:         host,
		csrSubject:   csrSubject,
	}
}

func (ih *infoHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		api.RespondWithError(w, apperrors.Forbidden("Token not provided."))
		return
	}

	identifier := mux.Vars(r)["identifier"]

	cachedToken, found := ih.tokenService.GetClusterToken(identifier)
	if !found || cachedToken != token {
		api.RespondWithError(w, apperrors.Forbidden("Invalid token."))
		return
	}

	newToken, err := ih.tokenService.CreateClusterToken(identifier)
	if err != nil {
		api.RespondWithError(w, apperrors.Internal("Failed to generate new token."))
		return
	}

	signUrl := fmt.Sprintf(SignUrl, ih.host, identifier, newToken)
	certInfo := api.MakeCertInfo(ih.csrSubject, identifier)

	api.RespondWithBody(w, http.StatusOK, InfoResponse{SignUrl: signUrl, CertificateInfo: certInfo})
}
