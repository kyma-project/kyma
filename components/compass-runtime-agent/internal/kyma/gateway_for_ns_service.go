package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	apiresources "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-ns"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

type gatewayForNamespaceService struct {
	applicationRepository applications.Repository
	converter             converters.Converter
	resourcesService      apiresources.Service
}

func NewGatewayForNamespaceService(applicationRepository applications.Repository, converter converters.Converter, resourcesService apiresources.Service) Service {
	return &gatewayForNamespaceService{
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *gatewayForNamespaceService) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	log.Infof("Applications passed to Sync gateway_for_ns_service: %d", len(directorApplications))

	currentApplications, err := s.getExistingRuntimeApplications()
	if err != nil {
		log.Errorf("Failed to get existing applications: %s.", err)
		return nil, err
	}

	compassCurrentApplications := s.filterCompassApplications(currentApplications)

	return s.apply(compassCurrentApplications, directorApplications), nil
}

func (s *gatewayForNamespaceService) apply(runtimeApplications []v1alpha1.Application, directorApplications []model.Application) []Result {
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

func (s *gatewayForNamespaceService) getExistingRuntimeApplications() ([]v1alpha1.Application, apperrors.AppError) {
	applications, err := s.applicationRepository.List(v1.ListOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to get application list: %s", err)
	}

	return applications.Items, nil
}

func (s *gatewayForNamespaceService) filterCompassApplications(applications []v1alpha1.Application) []v1alpha1.Application {
	var compassApplications []v1alpha1.Application

	for _, application := range applications {
		if application.Spec.CompassMetadata != nil {
			compassApplications = append(compassApplications, application)
		}
	}
	return compassApplications
}

func (s *gatewayForNamespaceService) createApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Infof("Creating applications.")
	results := make([]Result, 0)

	for _, directorApplication := range directorApplications {
		if !ApplicationExists(directorApplication.Name, runtimeApplications) {
			result := s.createApplication(directorApplication, s.converter.Do(directorApplication))
			results = append(results, result)
		}
	}

	return results
}

func (s *gatewayForNamespaceService) createApplication(directorApplication model.Application, runtimeApplication v1alpha1.Application) Result {
	log.Infof("Creating application '%s'.", directorApplication.Name)
	newRuntimeApplication, err := s.applicationRepository.Create(&runtimeApplication)
	if err != nil {
		log.Warningf("Failed to create application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	log.Infof("Creating API resources for application '%s'.", directorApplication.Name)
	err = s.createAPIResources(directorApplication, *newRuntimeApplication)
	if err != nil {
		log.Warningf("Failed to create API resources for application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	return newResult(runtimeApplication, directorApplication.ID, Create, nil)
}

func (s *gatewayForNamespaceService) createAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {
	return s.resourcesService.CreateAPIResources(directorApplication, runtimeApplication)
}

func (s *gatewayForNamespaceService) deleteApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Info("Deleting applications.")
	results := make([]Result, 0)

	for _, runtimeApplication := range runtimeApplications {
		existsInDirector := false
		for _, directorApp := range directorApplications {
			if directorApp.Name == runtimeApplication.Name {
				existsInDirector = true
				break
			}
		}

		if !existsInDirector {
			result := s.deleteApplication(runtimeApplication, runtimeApplication.GetApplicationID())
			results = append(results, result)
		}
	}
	return results
}

func (s *gatewayForNamespaceService) deleteApplication(runtimeApplication v1alpha1.Application, applicationID string) Result {
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

	return newResult(runtimeApplication, applicationID, Delete, err)
}

func (s *gatewayForNamespaceService) deleteAllAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {
	return s.resourcesService.DeleteAPIResources(runtimeApplication)
}

func (s *gatewayForNamespaceService) updateApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
	log.Info("Updating applications.")
	results := make([]Result, 0)

	for _, directorApplication := range directorApplications {
		if ApplicationExists(directorApplication.Name, runtimeApplications) {
			existentApplication := GetApplication(directorApplication.Name, runtimeApplications)
			result := s.updateApplication(directorApplication, existentApplication, s.converter.Do(directorApplication))
			results = append(results, result)
		}
	}

	return results
}

func (s *gatewayForNamespaceService) updateApplication(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) Result {
	log.Infof("Updating Application '%s'.", directorApplication.Name)
	updatedRuntimeApplication, err := s.applicationRepository.Update(&newRuntimeApplication)
	if err != nil {
		log.Warningf("Failed to update application '%s': %s.", directorApplication.Name, err)
		return newResult(existentRuntimeApplication, directorApplication.ID, Update, err)
	}

	log.Infof("Updating API resources for application '%s'.", directorApplication.Name)
	appendedErr := s.updateAPIResources(directorApplication, existentRuntimeApplication, *updatedRuntimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to update API resources for application '%s': %s.", directorApplication.Name, appendedErr)
	}

	return newResult(existentRuntimeApplication, directorApplication.ID, Update, appendedErr)
}

func (s *gatewayForNamespaceService) updateAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {

	appendedErr := s.resourcesService.UpsertAPIResources(directorApplication, existentRuntimeApplication, newRuntimeApplication)
	if appendedErr != nil {
		return appendedErr
	}

	return s.resourcesService.DeleteResourcesOfNonExistentAPI(existentRuntimeApplication, directorApplication, newRuntimeApplication.Name)
}
