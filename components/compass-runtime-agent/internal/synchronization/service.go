package synchronization

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []compass.Application) ([]Result, apperrors.AppError)
}

func NewSynchronizationService() Service {
	return &service{}
}

type service struct {
	reconciler            Reconciler
	applicationRepository ApplicationRepository
	converter             Converter
	resourcesService      ResourcesService
}

type Result struct {
	ApplicationID string
	Operation     Operation
	Error         apperrors.AppError
}

type Reconciler interface {
	Do(applications []compass.Application) ([]ApplicationAction, apperrors.AppError)
}

type ApplicationRepository interface {
	Create(application *v1alpha1.Application) apperrors.AppError
	Update(application *v1alpha1.Application) apperrors.AppError
	Delete(applicationID string) apperrors.AppError
}
type ApiIDToSecretNameMap map[string]string

type ResourcesService interface {
	CreateApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	CreateEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	CreateSecrets(application compass.Application, apiDefinition graphql.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	UpdateApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	UpdateEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	UpdateSecrets(application compass.Application, apiDefinition graphql.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	DeleteApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
	DeleteEventApiResources(application compass.Application, eventApiDefinition graphql.EventAPIDefinition) apperrors.AppError
	DeleteSecrets(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError
}

func NewService(reconciler Reconciler, applicationRepository ApplicationRepository, converter Converter, resourcesService ResourcesService) Service {
	return Service{
		reconciler:            reconciler,
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *service) Apply(applications []compass.Application) ([]Result, apperrors.AppError) {

	actions, err := s.reconciler.Do(applications)
	if err != nil {
		return nil, err
	}

	results := make([]Result, len(actions))

	for i, action := range actions {
		results[i] = s.apply(action)
	}

	return results, nil
}

func (s Service) apply(action ApplicationAction) Result {

	app := action.Application
	operation := action.Operation
	apiActions := action.APIActions
	eventActions := action.EventAPIActions

	var err apperrors.AppError

	switch action.Operation {
	case Create:
		err = s.applyCreateOperation(app, apiActions, eventActions)
	case Delete:
		err = s.applyDeleteOperation(app, apiActions, eventActions)
	case Update:
		err = s.applyUpdateOperation(app, apiActions, eventActions)
	}

	return newResult(app, operation, err)
}

func (s Service) applyCreateOperation(application compass.Application, apiActions []APIAction, eventAPIActions []EventAPIAction) apperrors.AppError {

	var err apperrors.AppError
	credentialSecretNames, paramsSecretNames, e := s.applyApiAndEventActions(application, apiActions, eventAPIActions)

	newApp := s.converter.Do(application)

	s.updateSecrets(&newApp, credentialSecretNames, paramsSecretNames)

	e = s.applicationRepository.Create(&newApp)
	if e != nil {
		err = appendError(err, e)
	}

	return err
}

func (s Service) applyUpdateOperation(application compass.Application, apiActions []APIAction, eventAPIActions []EventAPIAction) apperrors.AppError {
	var err apperrors.AppError

	credentialSecretNames, paramsSecretNames, e := s.applyApiAndEventActions(application, apiActions, eventAPIActions)
	newApp := s.converter.Do(application)

	s.updateSecrets(&newApp, credentialSecretNames, paramsSecretNames)
	e = s.applicationRepository.Update(&newApp)
	if e != nil {
		err = appendError(err, e)
	}

	return err
}

func (s Service) applyDeleteOperation(application compass.Application, apiActions []APIAction, eventAPIActions []EventAPIAction) apperrors.AppError {
	var err apperrors.AppError

	_, _, e := s.applyApiAndEventActions(application, apiActions, eventAPIActions)
	if e != nil {
		err = appendError(err, e)
	}

	e = s.applicationRepository.Delete(application.ID)
	if e != nil {
		err = appendError(err, e)
	}

	return err
}

func (s Service) applyApiAndEventActions(application compass.Application, apiActions []APIAction, eventAPIActions []EventAPIAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {
	err = s.applyApiResources(application, apiActions)

	e := s.applyEventResources(application, eventAPIActions)
	if e != nil {
		err = appendError(err, e)
	}

	credentials, params, e = s.applyApiSecrets(application, apiActions)
	if e != nil {
		err = appendError(err, e)
	}

	return credentials, params, err
}

func (s Service) applyApiResources(application compass.Application, apiActions []APIAction) apperrors.AppError {

	var err apperrors.AppError
	for _, apiAction := range apiActions {
		switch apiAction.Operation {
		case Create:
			e := s.resourcesService.CreateApiResources(application, apiAction.API)
			err = appendError(err, e)
		case Update:
			e := s.resourcesService.UpdateApiResources(application, apiAction.API)
			err = appendError(err, e)
		case Delete:
			e := s.resourcesService.DeleteApiResources(application, apiAction.API)
			err = appendError(err, e)
		}
	}

	return err
}

func (s Service) applyEventResources(application compass.Application, eventAPIActions []EventAPIAction) apperrors.AppError {
	var err apperrors.AppError
	for _, eventApiAction := range eventAPIActions {
		switch eventApiAction.Operation {
		case Create:
			e := s.resourcesService.CreateEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		case Update:
			e := s.resourcesService.UpdateEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		case Delete:
			e := s.resourcesService.DeleteEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		}
	}

	return err
}

func (s Service) applyApiSecrets(application compass.Application, APIActions []APIAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {

	credentials = make(map[string]string)
	params = make(map[string]string)

	for _, apiAction := range APIActions {
		switch apiAction.Operation {
		case Create:
			credSecretNames, paramsSecretNames, e := s.resourcesService.CreateSecrets(application, apiAction.API)
			appendMap(credentials, credSecretNames)
			appendMap(params, paramsSecretNames)
			err = appendError(err, e)
		case Update:
			secretNames, paramsSecretNames, e := s.resourcesService.UpdateSecrets(application, apiAction.API)
			appendMap(credentials, secretNames)
			appendMap(params, paramsSecretNames)
			err = appendError(err, e)
		case Delete:
			e := s.resourcesService.DeleteSecrets(application, apiAction.API)
			err = appendError(err, e)
		}
	}

	return credentials, params, err
}

func appendMap(target map[string]string, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}

func newResult(application compass.Application, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationID: application.ID,
		Operation:     operation,
		Error:         appError,
	}
}

// TODO: consider getting rid of this function and passing secrets data to converter instead
func (s Service) updateSecrets(application *v1alpha1.Application, credentialSecretNames ApiIDToSecretNameMap, paramsSecretNames ApiIDToSecretNameMap) {
	for _, service := range application.Spec.Services {
		for _, entry := range service.Entries {
			if entry.ApiType == specAPIType {
				secretName, found := credentialSecretNames[service.ID]
				if found {
					entry.Credentials.SecretName = secretName
				}

				secretName, found = paramsSecretNames[service.ID]
				if found {
					entry.RequestParametersSecretName = secretName
				}
			}
		}
	}
}

func appendError(wrapped apperrors.AppError, new apperrors.AppError) apperrors.AppError {
	if wrapped == nil {
		return new
	}

	return wrapped.Append("", new)
}
