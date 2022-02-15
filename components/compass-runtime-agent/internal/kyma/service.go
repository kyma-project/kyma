package kyma

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/rafter"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	appsecrets "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/secrets"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type service struct {
	applicationRepository    applications.Repository
	converter                applications.Converter
	rafter                   rafter.Service
	credentialsService       appsecrets.CredentialsService
	requestParametersService appsecrets.RequestParametersService
}

//go:generate mockery --name=Service
type Service interface {
	Apply(applications []model.Application) ([]Result, apperrors.AppError)
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

func NewService(applicationRepository applications.Repository, converter applications.Converter, resourcesService rafter.Service, credentialsService appsecrets.CredentialsService, requestParametersService appsecrets.RequestParametersService) Service {
	return &service{
		applicationRepository:    applicationRepository,
		converter:                converter,
		rafter:                   resourcesService,
		credentialsService:       credentialsService,
		requestParametersService: requestParametersService,
	}
}

func (s *service) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	log.Infof("Applications passed to Sync service: %d", len(directorApplications))

	currentApplications, err := s.getExistingRuntimeApplications()
	if err != nil {
		log.Errorf("Failed to get existing applications: %s.", err)
		return nil, err
	}

	compassCurrentApplications := s.filterCompassApplications(currentApplications)

	return s.apply(compassCurrentApplications, directorApplications), nil
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
	applicationList, err := s.applicationRepository.List(v1.ListOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to get application list: %s", err)
	}

	return applicationList.Items, nil
}

func (s *service) getApplicationUID(application string) (types.UID, apperrors.AppError) {
	app, err := s.applicationRepository.Get(application, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application %s not found", application)
			return "", apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Application %s failed, %s", application, err.Error())
		return "", apperrors.Internal(message)
	}

	return app.UID, nil
}

func (s *service) filterCompassApplications(applications []v1alpha1.Application) []v1alpha1.Application {
	var compassApplications []v1alpha1.Application

	for _, application := range applications {
		if application.Spec.CompassMetadata != nil {
			compassApplications = append(compassApplications, application)
		}
	}
	return compassApplications
}

func (s *service) createApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *service) createApplication(directorApplication model.Application, runtimeApplication v1alpha1.Application) Result {
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

	log.Infof("Creating credentials secrets for application '%s'.", directorApplication.Name)
	err = s.upsertCredentialsSecrets(directorApplication)
	if err != nil {
		log.Warningf("Failed to create credentials secrets for application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	log.Infof("Creating request parameters secrets for application '%s'.", directorApplication.Name)
	err = s.upsertRequestParametersSecrets(directorApplication)
	if err != nil {
		log.Warningf("Failed to create request parameters secrets for application '%s': %s.", directorApplication.Name, err)
		return newResult(runtimeApplication, directorApplication.ID, Create, err)
	}

	return newResult(runtimeApplication, directorApplication.ID, Create, nil)
}

func (s *service) upsertAPIResources(directorApplication model.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	for _, apiPackage := range directorApplication.APIPackages {
		err := s.upsertAPIResourcesForPackage(apiPackage)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	return appendedErr
}

func (s *service) upsertCredentialsSecrets(directorApplication model.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	getApplicationUIDFunc := cachingGetApplicationUIDFunc(s.getApplicationUID)
	for _, apiPackage := range directorApplication.APIPackages {
		if apiPackage.DefaultInstanceAuth != nil && apiPackage.DefaultInstanceAuth.Credentials != nil {
			credentials := apiPackage.DefaultInstanceAuth.Credentials
			if credentials.Basic == nil && credentials.Oauth == nil {
				continue
			}
			r, _ := getApplicationUIDFunc(directorApplication.Name)
			if r.AppError != nil {
				return r.AppError
			}
			_, err := s.credentialsService.Upsert(directorApplication.Name, r.AppUID, apiPackage.ID, credentials)
			if err != nil {
				appendedErr = apperrors.AppendError(appendedErr, err)
			}
		}
	}
	return appendedErr
}

func (s *service) upsertRequestParametersSecrets(directorApplication model.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	getApplicationUIDFunc := cachingGetApplicationUIDFunc(s.getApplicationUID)
	for _, apiPackage := range directorApplication.APIPackages {
		if apiPackage.DefaultInstanceAuth != nil && apiPackage.DefaultInstanceAuth.RequestParameters != nil && !apiPackage.DefaultInstanceAuth.RequestParameters.IsEmpty() {
			r, _ := getApplicationUIDFunc(directorApplication.Name)
			if r.AppError != nil {
				return r.AppError
			}
			requestParameters := apiPackage.DefaultInstanceAuth.RequestParameters
			if requestParameters != nil && !requestParameters.IsEmpty() {
				_, err := s.requestParametersService.Upsert(directorApplication.Name, r.AppUID, apiPackage.ID, requestParameters)
				if err != nil {
					appendedErr = apperrors.AppendError(appendedErr, err)
				}
			}
		}
	}
	return appendedErr
}

func (s *service) upsertAPIResourcesForPackage(apiPackage model.APIPackage) apperrors.AppError {
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

func (s *service) deleteApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *service) deleteApplication(runtimeApplication v1alpha1.Application, applicationID string) Result {

	log.Infof("Deleting request parameters secrets for application '%s'.", runtimeApplication.Name)
	if err := s.deleteRequestParametersSecrets(runtimeApplication); err != nil {
		log.Warningf("Failed to delete request parameters secrets secrets for application '%s': %s.", runtimeApplication.Name, err)
	}

	log.Infof("Deleting credentials secrets for application '%s'.", runtimeApplication.Name)
	if err := s.deleteCredentialsSecrets(runtimeApplication); err != nil {
		log.Warningf("Failed to delete credentials secrets for application '%s': %s.", runtimeApplication.Name, err)
	}

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

func (s *service) deleteAllAPIResources(runtimeApplication v1alpha1.Application) apperrors.AppError {
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

func (s *service) deleteCredentialsSecrets(runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	secretNames := s.getCredentialsSecretNames(runtimeApplication)

	for secretName := range secretNames {
		err := s.credentialsService.Delete(secretName)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}
	return appendedErr
}

func (s *service) getCredentialsSecretNames(runtimeApplication v1alpha1.Application) map[string]struct{} {
	secretNames := make(map[string]struct{})
	for _, service := range runtimeApplication.Spec.Services {
		for _, entry := range service.Entries {
			if entry.Credentials.SecretName != "" {
				secretNames[entry.Credentials.SecretName] = struct{}{}
			}
		}
	}
	return secretNames
}

func (s *service) deleteRequestParametersSecrets(runtimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	secretNames := s.getRequestParametersSecretNames(runtimeApplication)

	for secretName := range secretNames {
		err := s.requestParametersService.Delete(secretName)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}
	return appendedErr
}

func (s *service) getRequestParametersSecretNames(runtimeApplication v1alpha1.Application) map[string]struct{} {
	secretNames := make(map[string]struct{})
	for _, service := range runtimeApplication.Spec.Services {
		for _, entry := range service.Entries {
			if entry.RequestParametersSecretName != "" {
				secretNames[entry.RequestParametersSecretName] = struct{}{}
			}
		}
	}
	return secretNames
}

func (s *service) updateApplications(directorApplications []model.Application, runtimeApplications []v1alpha1.Application) []Result {
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

func (s *service) updateApplication(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) Result {
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

	log.Infof("Updating credentials secrets for application '%s'.", directorApplication.Name)
	appendedErr = s.updateCredentialsSecrets(directorApplication, existentRuntimeApplication, *updatedRuntimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to update credentials secrets for application '%s': %s.", directorApplication.Name, appendedErr)
	}

	log.Infof("Updating request paramters secrets for application '%s'.", directorApplication.Name)
	appendedErr = s.updateRequestParametersSecrets(directorApplication, existentRuntimeApplication, *updatedRuntimeApplication)
	if appendedErr != nil {
		log.Warningf("Failed to request paramters secrets for application '%s': %s.", directorApplication.Name, appendedErr)
	}

	return newResult(existentRuntimeApplication, directorApplication.ID, Update, appendedErr)
}

func (s *service) updateAPIResources(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {

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

func (s *service) updateCredentialsSecrets(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	// delete
	existentSecretNames := s.getCredentialsSecretNames(existentRuntimeApplication)
	newSecretNames := s.getCredentialsSecretNames(newRuntimeApplication)
	deletedSecretNames := make(map[string]struct{})
	for secretName := range existentSecretNames {
		if _, ok := newSecretNames[secretName]; !ok {
			deletedSecretNames[secretName] = struct{}{}
		}
	}
	for secretName := range deletedSecretNames {
		log.Infof("Deleting credentials secret '%s' for application '%s'", secretName, directorApplication.Name)
		err := s.credentialsService.Delete(secretName)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	// create + update
	err := s.upsertCredentialsSecrets(directorApplication)
	if err != nil {
		appendedErr = apperrors.AppendError(appendedErr, err)
	}
	return appendedErr
}

func (s *service) updateRequestParametersSecrets(directorApplication model.Application, existentRuntimeApplication v1alpha1.Application, newRuntimeApplication v1alpha1.Application) apperrors.AppError {
	var appendedErr apperrors.AppError

	// delete
	existentSecretNames := s.getRequestParametersSecretNames(existentRuntimeApplication)
	newSecretNames := s.getRequestParametersSecretNames(newRuntimeApplication)
	deletedSecretNames := make(map[string]struct{})
	for secretName := range existentSecretNames {
		if _, ok := newSecretNames[secretName]; !ok {
			deletedSecretNames[secretName] = struct{}{}
		}
	}
	for secretName := range deletedSecretNames {
		log.Infof("Deleting request parameters secret '%s' for application '%s'", secretName, directorApplication.Name)
		err := s.requestParametersService.Delete(secretName)
		if err != nil {
			appendedErr = apperrors.AppendError(appendedErr, err)
		}
	}

	// create + update
	err := s.upsertRequestParametersSecrets(directorApplication)
	if err != nil {
		appendedErr = apperrors.AppendError(appendedErr, err)
	}
	return appendedErr
}
