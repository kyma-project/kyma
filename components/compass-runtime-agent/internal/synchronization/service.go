package synchronization

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/docstopic"
	"k8s.io/apimachinery/pkg/types"
)

type Service struct {
	reconciler            Reconciler
	applicationRepository ApplicationRepository
	assetStoreService     AssetStore
	accessService         AccessResources
	converter             Converter
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
	Create(application v1alpha1.Application) apperrors.AppError
	Update(application v1alpha1.Application) apperrors.AppError
	Delete(application v1alpha1.Application) apperrors.AppError
}

type AssetStore interface {
	Create(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Update(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Delete(id string) apperrors.AppError
}

type AccessResources interface {
	Create(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Update(applicationName string, applicationUID types.UID, apiID, serviceName string) apperrors.AppError
	Delete(serviceName string) apperrors.AppError
}

type SecretResources interface {
	// TODO
}

func NewService(reconciler Reconciler, applicationRepository ApplicationRepository, converter Converter, assetStoreService AssetStore, accessService AccessResources) Service {
	return Service{
		reconciler:            reconciler,
		applicationRepository: applicationRepository,

		assetStoreService: assetStoreService,
		accessService:     accessService,
	}
}

func (s Service) Apply(applications []compass.Application) ([]Result, apperrors.AppError) {

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

	//app := action.Application
	//operation := action.Operation
	//apiActions := action.APIActions
	//eventActions := action.EventAPIActions
	//
	//var err apperrors.AppError
	//
	//switch action.Operation {
	//case Create:
	//	err = s.createApplication(app, apiActions, eventActions)
	//case Delete:
	//	err = s.deleteApplication(app, apiActions, eventActions)
	//case Update:
	//	err = s.updateApplication(app, apiActions, eventActions)
	//}
	//
	//return newResult(app, operation, err)

	return Result{}
}

func newResult(application compass.Application, operation Operation, appError apperrors.AppError) Result {
	return Result{
		ApplicationID: application.ID,
		Operation:     operation,
		Error:         appError,
	}
}

func (s Service) createApplication(application compass.Application, apiDefinitions []graphql.APIDefinition, eventAPIDefinition []graphql.EventAPIDefinition) apperrors.AppError {

	var err apperrors.AppError

	for _, api := range apiDefinitions {
		e := s.createApiResources(application, api)
		if e != nil {
			err = appendError(err, e)
		}
	}

	for _, eventAPI := range eventAPIDefinition {
		e := s.createEventResources(application, eventAPI)
		if e != nil {
			err = appendError(err, e)
		}
	}

	newApp := s.converter.Do(application)
	e := s.applicationRepository.Create(newApp)
	if e != nil {
		err = appendError(err, e)
	}

	return err
}

func appendError(wrapped apperrors.AppError, new apperrors.AppError) apperrors.AppError {
	if wrapped == nil {
		return new
	}

	return wrapped.Append("", new)
}

func (s Service) createApiResources(application compass.Application, apiDefinition graphql.APIDefinition) apperrors.AppError {
	return nil
}

type ApiIDToSecretNameMap map[string]string

func (s Service) createApiSecrets(application compass.Application, apiDefinition graphql.APIDefinition) (ApiIDToSecretNameMap, apperrors.AppError) {
	return nil, nil
}

func (s Service) createEventResources(application compass.Application, eventAPIDefinition graphql.EventAPIDefinition) apperrors.AppError {
	return nil
}

func (s Service) deleteApplication(application compass.Application, APIActions []APIAction, EventAPIActions []EventAPIAction) apperrors.AppError {
	return nil
}

func (s Service) updateApplication(application compass.Application, APIActions []APIAction, EventAPIActions []EventAPIAction) apperrors.AppError {
	return nil
}
