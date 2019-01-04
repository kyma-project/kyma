package externalapi

import (
	"net/http"

	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/internal/kymagroup"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
)

type signatureHandler struct {
	tokenService    tokens.ClusterService
	certService     certificates.Service
	host            string
	groupRepository kymagroup.Repository
}

func NewSignatureHandler(tokenService tokens.ClusterService, certService certificates.Service, host string, groupRepository kymagroup.Repository) SignatureHandler {

	return &signatureHandler{
		tokenService:    tokenService,
		certService:     certService,
		host:            host,
		groupRepository: groupRepository,
	}
}

func (sh *signatureHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		api.RespondWithError(w, apperrors.Forbidden("Token not provided."))
		return
	}

	identifier := mux.Vars(r)["identifier"]

	cachedToken, found := sh.tokenService.GetClusterToken(identifier)
	if !found || cachedToken != token {
		api.RespondWithError(w, apperrors.Forbidden("Invalid token."))
		return
	}

	var certificateRequest CertificateRequest
	err := api.ReadRequestBody(r, &certificateRequest)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	signedCrt, err := sh.certService.SignCSR(certificateRequest.CSR, identifier)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	err = sh.updateKymaGroupClusterData(identifier, &certificateRequest)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	sh.tokenService.DeleteClusterToken(identifier)

	api.RespondWithBody(w, http.StatusCreated, api.CertificateResponse{CRT: signedCrt})
}

// We assume that KymaGroup already exists when Kyma Cluster is registering
// If it is not the case operation will fail
func (sh *signatureHandler) updateKymaGroupClusterData(groupId string, certRequest *CertificateRequest) apperrors.AppError {
	if certRequest.KymaCluster.AppRegistryUrl == "" || certRequest.KymaCluster.EventsUrl == "" {
		return nil
	}

	clusterInfo := &v1alpha1.Cluster{
		AppRegistryUrl: certRequest.KymaCluster.AppRegistryUrl,
		EventsUrl:      certRequest.KymaCluster.EventsUrl,
	}

	return sh.groupRepository.UpdateClusterData(groupId, clusterInfo)
}
