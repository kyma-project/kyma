package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

type gatewayForNamespaceService struct {
	applicationRepository applications.Repository
	converter             converters.Converter
	rafter                rafter.Service
}

func NewGatewayForNsService(applicationRepository applications.Repository, converter converters.Converter, resourcesService rafter.Service) Service {
	return &gatewayForNamespaceService{
		applicationRepository: applicationRepository,
		converter:             converter,
		rafter:                resourcesService,
	}
}

func (s *gatewayForNamespaceService) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	log.Infof("Applications passed to Sync service: %d", len(directorApplications))

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
	_, err := s.applicationRepository.Create(&runtimeApplication)
	if err != nil {
		log.Warningf("Failed to create application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	log.Infof("Creating API resources for application '%s'.", directorApplication.Name)
	err = s.upsertAPIResources(directorApplication)
	if err != nil {
		log.Warningf("Failed to create API resources for application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	return newResult(runtimeApplication, directorApplication.ID, Create, nil)
}

func (s *gatewayForNamespaceService) upsertAPIResources(directorApplication model.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiPackage := range directorApplication.APIPackages {
		err := s.upsertAPIResourcesForPackage(apiPackage)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
}

func (s *gatewayForNamespaceService) upsertAPIResourcesForPackage(apiPackage model.APIPackage) apperrors.AppError {
	if !model.PackageContainsAnySpecs(apiPackage) {
		return nil
	}

	assetsCount := len(apiPackage.APIDefinitions) + len(apiPackage.EventDefinitions)
	assets := make([]clusterassetgroup.Asset, 0, assetsCount)

	for _, apiDefinition := range apiPackage.APIDefinitions {
		if apiDefinition.APISpec != nil {
			assets = append(assets, createAssetFromAPIDefinition(apiDefinition))
		}
	}

	for _, eventAPIDefinition := range apiPackage.EventDefinitions {
		if eventAPIDefinition.EventAPISpec != nil {
			assets = append(assets, createAssetFromEventAPIDefinition(eventAPIDefinition))
		}
	}

	return s.rafter.Put(apiPackage.ID, assets)
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
	var appendedErr apperrors.AppError
	for _, service := range runtimeApplication.Spec.Services {
		log.Infof("Deleting resources for API '%s' and application '%s'", service.ID, runtimeApplication.Name)
		err := s.rafter.Delete(service.ID)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
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

	appendedErr := s.upsertAPIResources(directorApplication)

	for _, service := range existentRuntimeApplication.Spec.Services {
		apiPackage, apiPackageExists := model.APIPackageExists(service.ID, directorApplication)
		deleteSpecs := (apiPackageExists && !model.PackageContainsAnySpecs(apiPackage)) || !apiPackageExists

		if deleteSpecs {
			log.Infof("Deleting resources for API '%s' and application '%s'", service.ID, directorApplication.Name)
			err := s.rafter.Delete(service.ID)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
}
