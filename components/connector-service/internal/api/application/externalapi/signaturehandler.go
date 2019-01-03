package externalapi

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1Apps "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/connector-service/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/connector-service/internal/applications"

	"github.com/kyma-project/kyma/components/connector-service/internal/kymagroup"

	"github.com/kyma-project/kyma/components/connector-service/internal/api"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
)

type signatureHandler struct {
	tokenService    tokens.ApplicationService
	certService     certificates.Service
	host            string
	groupRepository kymagroup.Repository
	appRepository   applications.Repository
}

func NewSignatureHandler(tokenService tokens.ApplicationService, certService certificates.Service, host string, groupRepository kymagroup.Repository, appRepository applications.Repository) SignatureHandler {

	return &signatureHandler{
		tokenService:    tokenService,
		host:            host,
		certService:     certService,
		groupRepository: groupRepository,
		appRepository:   appRepository,
	}
}

func (sh *signatureHandler) SignCSR(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		api.RespondWithError(w, apperrors.Forbidden("Token not provided."))
		return
	}

	identifier := mux.Vars(r)["identifier"]

	tokenData, found := sh.tokenService.GetAppToken(identifier)
	if !found || tokenData.Token != token {
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

	err = sh.handleResources(identifier, tokenData, &certificateRequest)
	if err != nil {
		api.RespondWithError(w, err)
		return
	}

	sh.tokenService.DeleteAppToken(identifier)

	api.RespondWithBody(w, http.StatusCreated, api.CertificateResponse{CRT: signedCrt})
}

func (sh *signatureHandler) handleResources(identifier string, tokenData *tokens.TokenData, certRequest *CertificateRequest) apperrors.AppError {
	// TODO - should we handle case if APP is created but creating Group fails?

	err := sh.handleApplicationCR(identifier, certRequest)
	if err != nil {
		return err
	}

	err = sh.handleGroupCR(identifier, tokenData)
	if err != nil {
		return err
	}

	return nil
}

// TODO - validate that - Tenant, Group, Display Name - is unique
func (sh *signatureHandler) handleGroupCR(appIdentifier string, tokenData *tokens.TokenData) apperrors.AppError {
	appGroupEntry := &v1alpha1.Application{
		ID: appIdentifier,
	}

	group, err := sh.groupRepository.Get(tokenData.Group)
	if err != nil {
		if err.IsNotFound() {
			err = sh.createKymaGroup(tokenData.Group, appGroupEntry)
			if err != nil {
				return apperrors.Internal("Failed to create new group %s, %s", tokenData.Group, err.Error())
			}
		} else {
			return apperrors.Internal("Failed to read group %s, %s", tokenData.Group, err.Error())
		}
	} else {
		// TODO - clarify what should happen when Application already present in the group
		err := sh.groupRepository.AddApplication(group.Name, appGroupEntry)
		if err != nil {
			if !err.IsAlreadyExists() {
				return apperrors.Internal("Failed to add Application %s to group %s, %s", appIdentifier, tokenData.Group, err.Error())
			}
			logrus.Infof("Application %s already in group %s", appIdentifier, tokenData.Group)
		}
	}

	return nil
}

func (sh *signatureHandler) createKymaGroup(groupId string, app *v1alpha1.Application) apperrors.AppError {
	newGroup := &v1alpha1.KymaGroup{
		TypeMeta:   v1.TypeMeta{Kind: "KymaGroup", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: groupId},
		Spec: v1alpha1.KymaGroupSpec{
			Applications: []v1alpha1.Application{*app},
			Cluster:      v1alpha1.Cluster{},
		},
	}

	return sh.groupRepository.Create(newGroup)
}

// TODO - Application Operator should delete entry when App is deleted
// TODO - determine what happens when app exists, is it updated?
func (sh *signatureHandler) handleApplicationCR(identifier string, certRequest *CertificateRequest) apperrors.AppError {
	_, err := sh.appRepository.Get(identifier)
	if err != nil {
		if err.IsNotFound() {
			// TODO - set display name after updating the Application CRD
			registeredApp := &v1alpha1Apps.Application{
				TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: v1alpha1Apps.SchemeGroupVersion.String()},
				ObjectMeta: v1.ObjectMeta{Name: identifier},
				Spec: v1alpha1Apps.ApplicationSpec{
					Description: certRequest.Application.Description,
					Labels:      certRequest.Application.Labels,
					Services:    []v1alpha1Apps.Service{},
				},
			}

			err := sh.appRepository.Create(registeredApp)
			if err != nil {
				return apperrors.Internal("Failed to create Application %s, %s", identifier, err.Error())
			}
		} else {
			return apperrors.Internal("Failed to read Application %s, %s", identifier, err.Error())
		}
	}

	return nil
}
