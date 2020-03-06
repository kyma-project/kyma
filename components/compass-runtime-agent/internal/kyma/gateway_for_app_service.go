package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	apiresources "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app"
	secretsmodel "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []model.Application) ([]Result, apperrors.AppError)
}

type gatewayForAppService struct {
	applicationRepository applications.Repository
	converter             converters.Converter
	resourcesService      apiresources.Service
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type Result struct {
	ApplicationName string
	ApplicationID   string
	Operation       Operation
	Error           apperrors.AppError
}

type ApiIDToSecretNameMap map[string]string

func NewService(applicationRepository applications.Repository, converter converters.Converter, resourcesService apiresources.Service) Service {
	return &gatewayForAppService{
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *gatewayForAppService) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	log.Infof("Applications passed to Sync gateway_for_app_service: %d", len(directorApplications))

	currentApplications, err := s.getExistingRuntimeApplications()
	if err != nil {
		log.Errorf("Failed to get existing applications: %s.", err)
		return nil, err
	}

	compassCurrentApplications := s.filterCompassApplications(currentApplications)

	return s.apply(compassCurrentApplications, directorApplications), nil
}

func (s *gatewayForAppService) apply(runtimeApplications []v1alpha1.Application, directorApplications []model.Application) []Result {
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

func (s *gatewayForAppService) getExistingRuntimeApplications() ([]v1alpha1.Application, apperrors.AppError) {
	applications, err := s.applicationRepository.List(v1.ListOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to get application list: %s", err)
	}

	return applications.Items, nil
}

func (s *gatewayForAppService) filterCompassApplications(applications []v1alpha1.Application) []v1alpha1.Application {
	var compassApplications []v1alpha1.Application

	for _, application := range applications {
		if application.Spec.CompassMetadata != nil {
			compassApplications = append(compassApplications, application)
		}
	}
	return compassApplications
}

func (s *gatewayForAppService) createApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *gatewayForAppService) createApplication(directorApplication model.Application, runtimeApplication v1alpha1.Application) Result {
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

func (s *gatewayForAppService) createAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		service := GetService(apiDefinition.ID, runtimeApplication)

		assets := createAssetsFromAPIDefinition(apiDefinition)
		err := s.resourcesService.CreateApiResources(runtimeApplication.Name, runtimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), assets)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	for _, eventApiDefinition := range directorApplication.EventAPIs {
		service := GetService(eventApiDefinition.ID, runtimeApplication)

		assets := createAssetsFromEventAPIDefinition(eventApiDefinition)

		err := s.resourcesService.CreateEventApiResources(runtimeApplication.Name, service.ID, assets)
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

func (s *gatewayForAppService) deleteApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *gatewayForAppService) deleteApplication(runtimeApplication v1alpha1.Application, applicationID string) Result {
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

func (s *gatewayForAppService) deleteAllAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {
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

func (s *gatewayForAppService) deleteAPIResources(applicationName string, service v1alpha1.Service) apperrors.AppError {
	for _, entry := range service.Entries {
		err := s.resourcesService.DeleteApiResources(applicationName, service.ID, entry.Credentials.SecretName)
		if err != nil {
			log.Warningf("Failed to delete resources for API '%s' and application '%s': %s", service.ID, service.Name, err)

			return err
		}
	}

	return nil
}

func (s *gatewayForAppService) updateApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *gatewayForAppService) updateApplication(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) Result {
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

func (s *gatewayForAppService) updateAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
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

func (s *gatewayForAppService) updateOrCreateRESTAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		existsInRuntime := ServiceExists(apiDefinition.ID, existentRuntimeApplication)
		service := GetService(apiDefinition.ID, newRuntimeApplication)

		assets := createAssetsFromAPIDefinition(apiDefinition)

		if existsInRuntime {
			log.Infof("Updating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.Name)
			err := s.resourcesService.UpdateApiResources(newRuntimeApplication.Name, newRuntimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), assets)
			if err != nil {
				log.Warningf("Failed to update API '%s': %s.", apiDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", apiDefinition.ID, directorApplication.Name)
			err := s.resourcesService.CreateApiResources(newRuntimeApplication.Name, newRuntimeApplication.UID, service.ID, toSecretsModel(apiDefinition.Credentials), assets)
			if err != nil {
				log.Warningf("Failed to create API '%s': %s.", apiDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		}
	}

	return appendedErr
}

func (s *gatewayForAppService) updateOrCreateEventAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, eventAPIDefinition := range directorApplication.EventAPIs {
		existsInRuntime := ServiceExists(eventAPIDefinition.ID, existentRuntimeApplication)
		service := GetService(eventAPIDefinition.ID, newRuntimeApplication)

		assets := []clusterassetgroup.Asset{
			{
				Name:    eventAPIDefinition.ID,
				Type:    getEventApiType(eventAPIDefinition.EventAPISpec),
				Content: getEventSpec(eventAPIDefinition.EventAPISpec),
				Format:  getEventSpecFormat(eventAPIDefinition.EventAPISpec),
			},
		}

		if existsInRuntime {
			log.Infof("Updating resources for API '%s' and application '%s'", eventAPIDefinition.ID, directorApplication.Name)

			err := s.resourcesService.UpdateEventApiResources(newRuntimeApplication.Name, service.ID, assets)
			if err != nil {
				log.Warningf("Failed to update Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		} else {
			log.Infof("Creating resources for API '%s' and application '%s'", eventAPIDefinition.ID, directorApplication.Name)

			err := s.resourcesService.CreateEventApiResources(newRuntimeApplication.Name, service.ID, assets)
			if err != nil {
				log.Warningf("Failed to create Event API '%s': %s.", eventAPIDefinition.ID, err)
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		}
	}

	return appendedErr
}

func (s *gatewayForAppService) deleteResourcesOfNonExistentAPI(existentRuntimeApplication v1alpha1.Application, directorApplication model.Application, name string) apperrors.AppError {
	var appendedErr apperrors.AppError
	for _, service := range existentRuntimeApplication.Spec.Services {
		if !model.APIExists(service.ID, directorApplication) {
			log.Infof("Deleting resources for API '%s' and application '%s'", service.ID, directorApplication.Name)
			err := s.deleteAPIResources(name, service)
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}
	return appendedErr
}
