package kyma

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []model.Application) ([]Result, apperrors.AppError)
}

func NewSynchronizationService() Service {
	return &service{}
}

type service struct {
	reconciler            sync.Reconciler
	applicationRepository applications.Manager
	converter             applications.Converter
	resourcesService      ResourcesService
}

type Result struct {
	ApplicationID string
	Operation     sync.Operation
	Error         apperrors.AppError
}

type ApiIDToSecretNameMap map[string]string

type ResourcesService interface {
	CreateApiResources(application model.Application, apiDefinition model.APIDefinition) apperrors.AppError
	CreateEventApiResources(application model.Application, eventApiDefinition model.EventAPIDefinition) apperrors.AppError
	CreateSecrets(application model.Application, apiDefinition model.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	UpdateApiResources(application model.Application, apiDefinition model.APIDefinition) apperrors.AppError
	UpdateEventApiResources(application model.Application, eventApiDefinition model.EventAPIDefinition) apperrors.AppError
	UpdateSecrets(application model.Application, apiDefinition model.APIDefinition) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError)

	DeleteApiResources(application model.Application, apiDefinition model.APIDefinition) apperrors.AppError
	DeleteEventApiResources(application model.Application, eventApiDefinition model.EventAPIDefinition) apperrors.AppError
	DeleteSecrets(application model.Application, apiDefinition model.APIDefinition) apperrors.AppError
}

func NewService(reconciler sync.Reconciler, applicationRepository applications.Manager, converter applications.Converter, resourcesService ResourcesService) Service {
	return &service{
		reconciler:            reconciler,
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *service) Apply(applications []model.Application) ([]Result, apperrors.AppError) {

	logrus.Info("Application passed to Sync service: ", len(applications))

	return nil, nil

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

func (s *service) apply(action sync.ApplicationAction) Result {

	app := action.Application
	operation := action.Operation
	apiActions := action.APIActions
	eventActions := action.EventAPIActions

	var err apperrors.AppError

	switch action.Operation {
	case sync.Create:
		err = s.applyCreateOperation(app, apiActions, eventActions)
	case sync.Delete:
		err = s.applyDeleteOperation(app, apiActions, eventActions)
	case sync.Update:
		err = s.applyUpdateOperation(app, apiActions, eventActions)
	}

	return newResult(app, operation, err)
}

func (s *service) applyCreateOperation(application model.Application, apiActions []sync.APIAction, eventAPIActions []sync.EventAPIAction) apperrors.AppError {

	var err apperrors.AppError
	s.applyApiAndEventActions(application, apiActions, eventAPIActions)

	newApp := s.converter.Do(application)

	{
		_, e := s.applicationRepository.Create(&newApp)
		if e != nil {
			err = appendError(err, apperrors.Internal("Failed to create Application: %s", e))
		}
	}

	return err
}

func (s *service) applyUpdateOperation(application model.Application, apiActions []sync.APIAction, eventAPIActions []sync.EventAPIAction) apperrors.AppError {
	var err apperrors.AppError

	s.applyApiAndEventActions(application, apiActions, eventAPIActions)
	newApp := s.converter.Do(application)

	{
		_, e := s.applicationRepository.Update(&newApp)
		if e != nil {
			err = appendError(err, apperrors.Internal("Failed to update application: %s", e))
		}
	}

	return err
}

func (s *service) applyDeleteOperation(application model.Application, apiActions []sync.APIAction, eventAPIActions []sync.EventAPIAction) apperrors.AppError {
	var err apperrors.AppError

	_, _, e := s.applyApiAndEventActions(application, apiActions, eventAPIActions)
	if e != nil {
		err = appendError(err, e)
	}

	{
		e := s.applicationRepository.Delete(application.Name, &v1.DeleteOptions{})
		if e != nil {
			err = appendError(err, apperrors.Internal("Failed to delete Application: %s", e))
		}
	}

	return err
}

func (s *service) applyApiAndEventActions(application model.Application, apiActions []sync.APIAction, eventAPIActions []sync.EventAPIAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {
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

func (s *service) applyApiResources(application model.Application, apiActions []sync.APIAction) apperrors.AppError {

	var err apperrors.AppError
	for _, apiAction := range apiActions {
		switch apiAction.Operation {
		case sync.Create:
			e := s.resourcesService.CreateApiResources(application, apiAction.API)
			err = appendError(err, e)
		case sync.Update:
			e := s.resourcesService.UpdateApiResources(application, apiAction.API)
			err = appendError(err, e)
		case sync.Delete:
			e := s.resourcesService.DeleteApiResources(application, apiAction.API)
			err = appendError(err, e)
		}
	}

	return err
}

func (s *service) applyEventResources(application model.Application, eventAPIActions []sync.EventAPIAction) apperrors.AppError {
	var err apperrors.AppError
	for _, eventApiAction := range eventAPIActions {
		switch eventApiAction.Operation {
		case sync.Create:
			e := s.resourcesService.CreateEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		case sync.Update:
			e := s.resourcesService.UpdateEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		case sync.Delete:
			e := s.resourcesService.DeleteEventApiResources(application, eventApiAction.EventAPI)
			err = appendError(err, e)
		}
	}

	return err
}

func (s *service) applyApiSecrets(application model.Application, APIActions []sync.APIAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {

	credentials = make(map[string]string)
	params = make(map[string]string)

	for _, apiAction := range APIActions {
		switch apiAction.Operation {
		case sync.Create:
			credSecretNames, paramsSecretNames, e := s.resourcesService.CreateSecrets(application, apiAction.API)
			appendMap(credentials, credSecretNames)
			appendMap(params, paramsSecretNames)
			err = appendError(err, e)
		case sync.Update:
			secretNames, paramsSecretNames, e := s.resourcesService.UpdateSecrets(application, apiAction.API)
			appendMap(credentials, secretNames)
			appendMap(params, paramsSecretNames)
			err = appendError(err, e)
		case sync.Delete:
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

func newResult(application model.Application, operation sync.Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationID: application.ID,
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
