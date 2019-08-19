package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	secretsmodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []model.Application) ([]Result, apperrors.AppError)
}

type service struct {
	applicationRepository applications.Repository
	converter             applications.Converter
	resourcesService      apiresources.Service
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type Result struct {
	ApplicationID string
	Operation     Operation
	Error         apperrors.AppError
}

type ApiIDToSecretNameMap map[string]string

func NewService(applicationRepository applications.Repository, converter applications.Converter, resourcesService apiresources.Service) Service {
	return &service{
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *service) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	log.Infof("Applications passed to Sync service: %d", len(directorApplications))

	currentApplications, err := s.getExistingRuntimeApplications()
	if err != nil {
		log.Errorf("Failed to get existing applications: %s.", err)
		return nil, err
	}

	return s.apply(currentApplications, directorApplications), nil
}

func (s *service) apply(runtimeApplications []v1alpha1.Application, directorApplications []model.Application) []Result {
	log.Infof("Applying configuration from the Compass Director.")
	results := make([]Result, 0)

	created := s.createApplications(directorApplications, runtimeApplications)
	deleted := s.deleteApplications(directorApplications, runtimeApplications)
	updated := s.updateApplications(directorApplications, runtimeApplications)

	results = append(results, created...)
	results = append(results, deleted...)
	results = append(results, updated...)

	return results
}

func (s *service) getExistingRuntimeApplications() ([]v1alpha1.Application, apperrors.AppError) {
	applications, err := s.applicationRepository.List(v1.ListOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to get application list: %s", err)
	}

	return applications.Items, nil
}

func (s *service) createApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Infof("Creating applications.")
	results := make([]Result, 0)

	for _, directorApplication := range directorApplications {
		if !applications.ApplicationExists(directorApplication.ID, runtimeApplications) {
			result := s.createApplication(directorApplication, s.converter.Do(directorApplication))
			results = append(results, result)
		}
	}

	return results
}

func (s *service) createApplication(directorApplication model.Application, runtimeApplication v1alpha1.Application) Result {
	log.Infof("Creating application '%s'.", directorApplication.ID)
	newRuntimeApplication, err := s.applicationRepository.Create(&runtimeApplication)
	if err != nil {
		log.Warningf("Failed to create application '%s': %s.", directorApplication.ID, err)
		return newResult(runtimeApplication, Create, err)
	}

	log.Infof("Creating API resources for application '%s'.", directorApplication.ID)
	err = s.createAPIResources(directorApplication, *newRuntimeApplication)
	if err != nil {
		log.Warningf("Failed to create API resources for application '%s': %s.", directorApplication.ID, err)
		return newResult(runtimeApplication, Create, err)
	}

	return newResult(runtimeApplication, Create, nil)
}

func (s *service) createAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		spec := getSpec(apiDefinition.APISpec)
		apiType := getApiType(apiDefinition.APISpec)
		service := applications.GetService(apiDefinition.ID, runtimeApplication)

		err := s.resourcesService.CreateApiResources(runtimeApplication.Name, runtimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), spec, apiType)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	for _, eventApiDefinition := range directorApplication.EventAPIs {
		spec := getEventSpec(eventApiDefinition.EventAPISpec)
		apiType := getEventApiType(eventApiDefinition.EventAPISpec)
		service := applications.GetService(eventApiDefinition.ID, runtimeApplication)

		err := s.resourcesService.CreateEventApiResources(runtimeApplication.Name, service.ID, spec, apiType)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
}

func toSecretsModel(credentials *model.Credentials) *secretsmodel.CredentialsWithCSRF {

	toCSRF := func(csrfInfo *model.CSRFInfo) *secretsmodel.CSRFInfo {
		if csrfInfo == nil {
			return nil
		}

		return &secretsmodel.CSRFInfo{
			TokenEndpointURL: csrfInfo.TokenEndpointURL,
		}

	}

	if credentials != nil && credentials.Basic != nil {
		return &secretsmodel.CredentialsWithCSRF{
			Basic: &secretsmodel.Basic{
				Username: credentials.Basic.Username,
				Password: credentials.Basic.Password,
			},
			CSRFInfo: toCSRF(credentials.CSRFInfo),
		}
	}

	if credentials != nil && credentials.Oauth != nil {
		return &secretsmodel.CredentialsWithCSRF{
			Oauth: &secretsmodel.Oauth{
				ClientID:     credentials.Oauth.ClientID,
				ClientSecret: credentials.Oauth.ClientSecret,
			},
			CSRFInfo: toCSRF(credentials.CSRFInfo),
		}
	}

	return nil
}

func (s *service) deleteApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Info("Deleting applications.")
	results := make([]Result, 0)

	for _, runtimeApplication := range runtimeApplications {
		existsInDirector := false
		for _, directorApp := range directorApplications {
			if directorApp.ID == runtimeApplication.Name {
				existsInDirector = true
				break
			}
		}

		if !existsInDirector {
			result := s.deleteApplication(runtimeApplication)
			results = append(results, result)
		}
	}
	return results
}

func (s *service) deleteApplication(runtimeApplication v1alpha1.Application) Result {
	log.Infof("Deleting API resources for application '%s'.", runtimeApplication.Name)
	appendedErr := s.deleteAllAPIResources(runtimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to delete API resources for application '%s'.", runtimeApplication.Name)
	}

	log.Infof("Deleting application '%s'.", runtimeApplication.Name)
	err := s.applicationRepository.Delete(runtimeApplication.Name, &v1.DeleteOptions{})
	if err != nil {
		log.Warningf("Failed to delete application '%s'", runtimeApplication.Name)
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	return newResult(runtimeApplication, Delete, err)
}

func (s *service) deleteAllAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, runtimeService := range runtimeApplication.Spec.Services {
		log.Infof("Deleting resources for API '%s' and application '%s'", runtimeService.ID, runtimeApplication.Name)
		err := s.deleteAPIResources(runtimeApplication.Name, runtimeService)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
}

func (s *service) deleteAPIResources(applicationName string, service v1alpha1.Service) apperrors.AppError {
	for _, entry := range service.Entries {
		err := s.resourcesService.DeleteApiResources(applicationName, service.ID, entry.Credentials.SecretName)
		if err != nil {
			log.Warningf("Failed to delete resources for API '%s' and application '%s': %s", service.ID, service.Name, err)

			return err
		}
	}

	return nil
}

func (s *service) updateApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Info("Updating applications.")
	results := make([]Result, 0)

	for _, directorApplication := range directorApplications {
		if applications.ApplicationExists(directorApplication.ID, runtimeApplications) {
			existentApplication := applications.GetApplication(directorApplication.ID, runtimeApplications)
			result := s.updateApplication(directorApplication, existentApplication, s.converter.Do(directorApplication))
			results = append(results, result)
		}
	}

	return results
}

func (s *service) updateApplication(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) Result {
	log.Infof("Updating API resources for application '%s'.", directorApplication.ID)
	updatedRuntimeApplication, err := s.applicationRepository.Update(&newRuntimeApplication)
	if err != nil {
		log.Warningf("Failed to update application '%s': %s.", directorApplication.ID, err)
		return newResult(existentRuntimeApplication, Update, err)
	}

	log.Infof("Updating API resources for application '%s'.", directorApplication.ID)
	appendedErr := s.updateAPIResources(directorApplication, existentRuntimeApplication, *updatedRuntimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to update API resources for application '%s': %s.", directorApplication.ID, appendedErr)
	}

	return newResult(existentRuntimeApplication, Update, appendedErr)
}

func (s *service) updateAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	appendedErr := s.updateOrCreateRESTAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)

	err := s.updateOrCreateEventAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)
	if err != nil {
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	err = s.deleteResourcesOfNonExistentAPI(existentRuntimeApplication, directorApplication, newRuntimeApplication.Name)
	if err != nil {
		appendedErr = apperrors.AppendError(appendedErr, err)
	}

	return appendedErr
}

func (s *service) updateOrCreateRESTAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		existsInRuntime := applications.ServiceExists(apiDefinition.ID, existentRuntimeApplication)
		service := applications.GetService(apiDefinition.ID, newRuntimeApplication)

		if existsInRuntime {
			log.Infof("Updating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.ID)
			err := s.resourcesService.UpdateApiResources(newRuntimeApplication.Name, newRuntimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), getSpec(apiDefinition.APISpec), getApiType(apiDefinition.APISpec))
			if err != nil {
				log.Warningf("Failed to update API '%s': %s.", apiDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.ID)
			err := s.resourcesService.CreateApiResources(newRuntimeApplication.Name, newRuntimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), getSpec(apiDefinition.APISpec), getApiType(apiDefinition.APISpec))
			if err != nil {
				log.Warningf("Failed to create API '%s': %s.", apiDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		}
	}

	return appendedErr
}

func (s *service) updateOrCreateEventAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, eventAPIDefinition := range directorApplication.EventAPIs {
		existsInRuntime := applications.ServiceExists(eventAPIDefinition.ID, existentRuntimeApplication)
		service := applications.GetService(eventAPIDefinition.ID, newRuntimeApplication)

		if existsInRuntime {
			log.Infof("Updating resources for API '%s' and application '%s'", eventAPIDefinition.ID, directorApplication.ID)
			err := s.resourcesService.UpdateEventApiResources(newRuntimeApplication.Name, service.ID, getEventSpec(eventAPIDefinition.EventAPISpec), getEventApiType(eventAPIDefinition.EventAPISpec))
			if err != nil {
				log.Warningf("Failed to update Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", eventAPIDefinition.ID, directorApplication.ID)
			err := s.resourcesService.CreateEventApiResources(newRuntimeApplication.Name, service.ID, getEventSpec(eventAPIDefinition.EventAPISpec), getEventApiType(eventAPIDefinition.EventAPISpec))
			if err != nil {
				log.Warningf("Failed to create Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		}
	}

	return appendedErr
}

func (s *service) deleteResourcesOfNonExistentAPI(existentRuntimeApplication v1alpha1.Application, directorApplication model.Application, name string) apperrors.AppError {
	var appendedErr apperrors.AppError
	for _, service := range existentRuntimeApplication.Spec.Services {
		if !model.APIExists(service.ID, directorApplication) {
			log.Infof("Deleting resources for API '%s' and application '%s'", service.ID, directorApplication.ID)
			err := s.deleteAPIResources(name, service)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}
	return appendedErr
}

func getSpec(apiSpec *model.APISpec) []byte {
	if apiSpec == nil {
		return nil
	}

	return apiSpec.Data
}

func getEventSpec(eventApiSpec *model.EventAPISpec) []byte {
	if eventApiSpec == nil {
		return nil
	}

	return eventApiSpec.Data
}

func getApiType(apiSpec *model.APISpec) docstopic.ApiType {
	if apiSpec == nil {
		return docstopic.Empty
	}
	if apiSpec.Type == model.APISpecTypeOdata {
		return docstopic.ODataApiType
	}
	if apiSpec.Type == model.APISpecTypeOpenAPI {
		return docstopic.OpenApiType
	}
	return docstopic.Empty
}

func getEventApiType(eventApiSpec *model.EventAPISpec) docstopic.ApiType {
	if eventApiSpec == nil {
		return docstopic.Empty
	}
	if eventApiSpec.Type == model.EventAPISpecTypeAsyncAPI {
		return docstopic.AsyncApi
	}
	return docstopic.Empty
}

func newResult(application v1alpha1.Application, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationID: application.Name,
		Operation:     operation,
		Error:         appError,
	}
}
