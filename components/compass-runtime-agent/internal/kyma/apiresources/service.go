package apiresources

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/accessservice"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/istio"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
)

type ApiIDToSecretNameMap map[string]string

//go:generate mockery -name=Service
type Service interface {
	CreateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, assets []clusterassetgroup.Asset) apperrors.AppError
	CreateEventApiResources(applicationName string, serviceID string, assets []clusterassetgroup.Asset) apperrors.AppError
	UpdateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, assets []clusterassetgroup.Asset) apperrors.AppError
	UpdateEventApiResources(applicationName string, serviceID string, assets []clusterassetgroup.Asset) apperrors.AppError
	DeleteApiResources(applicationName string, serviceID string, secretName string) apperrors.AppError
}

type service struct {
	accessServiceManager      accessservice.AccessServiceManager
	secretsService            secrets.Service
	requestParameteresService secrets.RequestParametersService
	istioService              istio.Service
	nameResolver              k8sconsts.NameResolver
	rafter                    rafter.Service
}

func NewService(accessServiceManager accessservice.AccessServiceManager, secretsService secrets.Service, nameResolver k8sconsts.NameResolver, istioService istio.Service, rafter rafter.Service) Service {
	return service{
		accessServiceManager: accessServiceManager,
		secretsService:       secretsService,
		nameResolver:         nameResolver,
		istioService:         istioService,
		rafter:               rafter,
	}
}

type AccessServiceManager interface {
	Create(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Upsert(application string, appUID types.UID, serviceId, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

func (s service) CreateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, assets []clusterassetgroup.Asset) apperrors.AppError {
	k8sResourceName := s.nameResolver.GetResourceName(applicationName, serviceID)
	log.Infof("Creating access service '%s' for application '%s' and service '%s'.", k8sResourceName, applicationName, serviceID)
	appendedErr := s.accessServiceManager.Create(applicationName, applicationUID, serviceID, k8sResourceName)
	if appendedErr != nil {
		log.Infof("Failed to create access service for application '%s' and service '%s': %s.", applicationName, serviceID, appendedErr)
	}

	if credentials != nil {
		log.Infof("Creating secret for application '%s' and service '%s'.", applicationName, serviceID)
		err := s.secretsService.Create(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			log.Infof("Failed to create secret for application '%s' and service '%s': %s.", applicationName, serviceID, err)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	} else {
		log.Infof("Credentials for application '%s' and service '%s' not provided.", applicationName, serviceID)
	}

	err := s.istioService.Create(applicationName, applicationUID, serviceID, k8sResourceName)
	log.Infof("Creating istio resources for application '%s' and service '%s'.", applicationName, serviceID)

	if err != nil {
		log.Infof("Failed to create istio resources for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	err = s.rafter.Put(serviceID, assets)
	log.Infof("Uploading Api Spec for application '%s' and service '%s'.", applicationName, serviceID)

	if err != nil {
		log.Infof("Failed to upload Api Spec for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	return appendedErr
}

func (s service) CreateEventApiResources(applicationName string, serviceID string, assets []clusterassetgroup.Asset) apperrors.AppError {
	err := s.rafter.Put(serviceID, assets)
	log.Infof("Uploading Event Api Spec for application '%s' and service '%s'.", applicationName, serviceID)

	if err != nil {
		log.Infof("Failed to upload Event Api Spec for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		return err
	}

	return nil
}

func (s service) UpdateApiResources(applicationName string, applicationUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF, assets []clusterassetgroup.Asset) apperrors.AppError {
	k8sResourceName := s.nameResolver.GetResourceName(applicationName, serviceID)
	log.Infof("Updating access service '%s' for application '%s' and service '%s'.", k8sResourceName, applicationName, serviceID)
	appendedErr := s.accessServiceManager.Upsert(applicationName, applicationUID, serviceID, k8sResourceName)
	if appendedErr != nil {
		log.Infof("Failed to update access service for application '%s' and service '%s': %s.", applicationName, serviceID, appendedErr)
	}

	if credentials != nil {
		log.Infof("Updating secret for application '%s' and service '%s'.", applicationName, serviceID)
		err := s.secretsService.Upsert(applicationName, applicationUID, serviceID, credentials)
		if err != nil {
			log.Infof("Failed to update secret for application '%s' and service '%s': %s.", applicationName, serviceID, err)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	} else {
		log.Infof("Credentials for application '%s' and service '%s' not provided.", applicationName, serviceID)
		log.Infof("Deleting old secret for application '%s' and service '%s'.", applicationName, serviceID)
		secretName := s.nameResolver.GetCredentialsSecretName(applicationName, serviceID)

		err := s.secretsService.Delete(secretName)
		if err != nil {
			log.Infof("Failed to delete secret for application '%s' and service '%s': %s.", applicationName, serviceID, err)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	log.Infof("Updating istio resources for application '%s' and service '%s'.", applicationName, serviceID)
	appError := s.istioService.Upsert(applicationName, applicationUID, serviceID, k8sResourceName)

	if appError != nil {
		log.Infof("Failed to update istio resources for application '%s' and service '%s': %s.", applicationName, serviceID, appError)
		appendedErr = appendedErr.Append("", appError)
	}

	log.Infof("Updating Api Spec for application '%s' and service '%s'.", applicationName, serviceID)
	err := s.rafter.Put(serviceID, assets)

	if err != nil {
		log.Infof("Failed to update Api Spec for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	return appendedErr
}

func (s service) UpdateEventApiResources(applicationName string, serviceID string, assets []clusterassetgroup.Asset) apperrors.AppError {
	err := s.rafter.Put(serviceID, assets)
	log.Infof("Updating Api Spec for application '%s' and service '%s'.", applicationName, serviceID)

	if err != nil {
		log.Infof("Failed to update Api Spec for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		return err
	}
	return nil
}

func (s service) DeleteApiResources(applicationName string, serviceID string, secretName string) apperrors.AppError {
	k8sResourceName := s.nameResolver.GetResourceName(applicationName, serviceID)
	log.Infof("Deleting access service '%s' for application '%s' and service '%s'.", k8sResourceName, applicationName, serviceID)
	appendedErr := s.accessServiceManager.Delete(k8sResourceName)

	if secretName != "" {
		log.Infof("Deleting secret for application '%s' and service '%s'.", applicationName, serviceID)
		err := s.secretsService.Delete(secretName)
		if err != nil {
			log.Infof("Failed to delete secret for application '%s' and service '%s': %s.", applicationName, serviceID, err)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	} else {
		log.Infof("Credentials for application '%s' and service '%s' not provided.", applicationName, serviceID)
	}

	appError := s.istioService.Delete(k8sResourceName)
	log.Infof("Updating istio resources for application '%s' and service '%s'.", applicationName, serviceID)

	if appError != nil {
		log.Infof("Failed to update istio resources for application '%s' and service '%s': %s.", applicationName, serviceID, appError)
		appendedErr = appendedErr.Append("", appError)
	}

	err := s.rafter.Delete(serviceID)
	log.Infof("Removing Api Spec for application '%s' and service '%s'.", applicationName, serviceID)

	if err != nil {
		log.Infof("Failed to remove Api Spec for application '%s' and service '%s': %s.", applicationName, serviceID, err)
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	return appendedErr
}
