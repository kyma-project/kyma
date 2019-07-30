package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
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
	applicationRepository applications.Manager
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

func NewSynchronizationService() Service {
	return &service{}
}

func NewService(applicationRepository applications.Manager, converter applications.Converter, resourcesService apiresources.Service) Service {
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
	log.Infof("Creating API resources for application '%s'.", directorApplication.ID)
	appendedErr := s.createAPIResources(directorApplication, runtimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to create API resources for application '%s': %s.", directorApplication.ID, appendedErr)
	}

	log.Infof("Creating application '%s'.", directorApplication.ID)
	_, err := s.applicationRepository.Create(&runtimeApplication)
	if err != nil {
		log.Warningf("Failed to create application '%s': %s.", directorApplication.ID, err)
		appendedErr = appendError(appendedErr, apperrors.Internal("Failed to create application: '%s'", err))
	}

	return newResult(runtimeApplication, Create, appendedErr)
}

func (s *service) createAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		spec := getSpec(apiDefinition.APISpec)
		service := applications.GetService(apiDefinition.ID, runtimeApplication)

		err := s.resourcesService.CreateApiResources(runtimeApplication, service, spec)
		if err != nil {
			appendedErr = appendError(appendedErr, err)
		}
	}

	for _, eventApiDefinition := range directorApplication.EventAPIs {
		spec := getEventSpec(eventApiDefinition.EventAPISpec)
		service := applications.GetService(eventApiDefinition.ID, runtimeApplication)

		err := s.resourcesService.CreateApiResources(runtimeApplication, service, spec)
		if err != nil {
			appendedErr = appendError(appendedErr, err)
		}
	}

	return appendedErr
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
	err := s.deleteAPIResources(runtimeApplication)
	if err != nil {
		log.Warningf("Failed to delete API resources for application '%s'.", runtimeApplication.Name)
	}

	log.Infof("Deleting application '%s'.", runtimeApplication.Name)
	e := s.applicationRepository.Delete(runtimeApplication.Name, &v1.DeleteOptions{})
	if e != nil {
		log.Warningf("Failed to delete application '%s'", runtimeApplication.Name)
		err = appendError(err, apperrors.Internal("Failed to delete application: %s", e))
	}

	return newResult(runtimeApplication, Delete, err)
}

func (s *service) deleteAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, runtimeService := range runtimeApplication.Spec.Services {
		log.Infof("Deleting resources for API '%s' and application '%s'", runtimeService.ID, runtimeApplication.Name)

		err := s.resourcesService.DeleteApiResources(runtimeApplication, runtimeService)
		if err != nil {
			log.Warningf("Failed to delete resources for API '%s' and application '%s': %s", runtimeService.ID, runtimeApplication.Name, err)
			appendedErr = appendError(appendedErr, err)
		}
	}

	return appendedErr
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
	appendedErr := s.updateAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to update API resources for application '%s': %s.", directorApplication.ID, appendedErr)
	}

	log.Infof("Updating API resources for application '%s'.", directorApplication.ID)
	_, err := s.applicationRepository.Update(&newRuntimeApplication)
	if err != nil {
		log.Warningf("Failed to update application '%s': %s.", directorApplication.ID, err)
		appendedErr = appendError(appendedErr, apperrors.Internal("Failed to update application: %s.", err))
	}

	return newResult(existentRuntimeApplication, Update, appendedErr)
}

func (s *service) updateAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	appendedErr := s.updateOrCreateRESTAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)

	err := s.updateOrCreateEventAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)
	if err != nil {
		appendedErr = appendError(appendedErr, err)
	}

	for _, service := range existentRuntimeApplication.Spec.Services {
		found := false
		for _, apiDefinition := range directorApplication.APIs {
			if apiDefinition.ID == service.ID {
				found = true
				break
			}
		}

		for _, eventAPIDefinition := range directorApplication.EventAPIs {
			if eventAPIDefinition.ID == service.ID {
				found = true
				break
			}
		}

		if !found {
			log.Infof("Deleting resources for API '%s' and application '%s'", service.ID, directorApplication.ID)
			e := s.resourcesService.DeleteApiResources(existentRuntimeApplication, service)
			if e != nil {
				log.Warningf("Failed to delete API '%s': %s.", service.ID, e)
				err = appendError(err, e)
			}
		}
	}

	return appendedErr
}

func (s *service) updateOrCreateRESTAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		existsInRuntime := applications.ServiceExists(apiDefinition.ID, existentRuntimeApplication)

		if existsInRuntime {
			log.Infof("Updating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.ID)
			service := applications.GetService(apiDefinition.ID, existentRuntimeApplication)
			err := s.resourcesService.UpdateApiResources(newRuntimeApplication, service, getSpec(apiDefinition.APISpec))
			if err != nil {
				log.Warningf("Failed to update API '%s': %s.", apiDefinition.ID, err)
				appendedErr = appendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.ID)
			service := applications.GetService(apiDefinition.ID, newRuntimeApplication)
			err := s.resourcesService.CreateApiResources(newRuntimeApplication, service, getSpec(apiDefinition.APISpec))
			if err != nil {
				log.Warningf("Failed to create API '%s': %s.", apiDefinition.ID, err)
				appendedErr = appendError(appendedErr, err)
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
			err := s.resourcesService.UpdateApiResources(newRuntimeApplication, service, getEventSpec(eventAPIDefinition.EventAPISpec))
			if err != nil {
				log.Warningf("Failed to update Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = appendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", eventAPIDefinition.ID, directorApplication.ID)
			err := s.resourcesService.CreateApiResources(newRuntimeApplication, service, getEventSpec(eventAPIDefinition.EventAPISpec))
			if err != nil {
				log.Warningf("Failed to create Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = appendError(appendedErr, err)
			}
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

func newResult(application v1alpha1.Application, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationID: application.Name,
		Operation:     operation,
		Error:         appError,
	}
}

func appendError(wrapped apperrors.AppError, new apperrors.AppError) apperrors.AppError {
	if wrapped == nil {
		return new
	}

	return wrapped.Append("", new)
}
