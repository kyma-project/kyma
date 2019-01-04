package externalapi

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/internal/kymagroup"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/gorilla/mux"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
)

const (
	CertUrl = "https://%s/v1/applications/%s"
	SignUrl = "https://%s/v1/applications/%s/client-certs?token=%s"
)

type infoHandler struct {
	tokenService    tokens.ApplicationService
	host            string
	csrSubject      certificates.CSRSubject
	groupRepository kymagroup.Repository
}

func NewInfoHandler(tokenService tokens.ApplicationService, host string, csrSubject certificates.CSRSubject, groupRepository kymagroup.Repository) InfoHandler {
	return &infoHandler{
		tokenService:    tokenService,
		host:            host,
		csrSubject:      csrSubject,
		groupRepository: groupRepository,
	}
}

func (ih *infoHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		api.RespondWithError(w, apperrors.Forbidden("Token not provided."))
		return
	}

	identifier := mux.Vars(r)["identifier"]

	tokenData, found := ih.tokenService.GetAppToken(identifier)
	if !found || tokenData.Token != token {
		api.RespondWithError(w, apperrors.Forbidden("Invalid token."))
		return
	}

	certUrl := fmt.Sprintf(CertUrl, ih.host, identifier)
	apiData, err := ih.createApiData(certUrl, tokenData, identifier)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	newToken, err := ih.tokenService.CreateAppToken(identifier, tokenData)
	if err != nil {
		api.RespondWithError(w, apperrors.Internal("Failed to generate new token."))
		return
	}

	signUrl := fmt.Sprintf(SignUrl, ih.host, identifier, newToken)
	certInfo := api.MakeCertInfo(ih.csrSubject, identifier)

	api.RespondWithBody(w, 200, InfoResponse{SignUrl: signUrl, Api: apiData, CertificateInfo: certInfo})
}

func (ih *infoHandler) createApiData(certUrl string, tokenData *tokens.TokenData, identifier string) (Api, apperrors.AppError) {
	registryUrl := ""
	eventsUrl := ""

	group, err := ih.groupRepository.Get(tokenData.Group)
	if err == nil {
		registryUrl = fmt.Sprintf(group.Spec.Cluster.AppRegistryUrl, identifier)
		eventsUrl = fmt.Sprintf(group.Spec.Cluster.EventsUrl, identifier)
	} else {
		if err.Code() != apperrors.CodeNotFound {
			return Api{}, err.Append("Failed to read Cluster URLs")
		}
	}

	return Api{
		MetadataURL:     registryUrl,
		EventsURL:       eventsUrl,
		CertificatesUrl: certUrl,
	}, nil
}
