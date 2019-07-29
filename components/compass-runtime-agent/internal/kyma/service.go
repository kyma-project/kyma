package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
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
	resourcesService      apiresources.Service
}

type Result struct {
	ApplicationID string
	Operation     sync.Operation
	Error         apperrors.AppError
}

type ApiIDToSecretNameMap map[string]string

func NewService(reconciler sync.Reconciler, applicationRepository applications.Manager, converter applications.Converter, resourcesService apiresources.Service) Service {
	return &service{
		reconciler:            reconciler,
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *service) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {

	logrus.Info("Application passed to Sync service: ", len(directorApplications))

	applications := make([]v1alpha1.Application, 0, len(directorApplications))

	for _, directorApplication := range directorApplications {
		applications = append(applications, s.converter.Do(directorApplication))
	}

	results := make([]Result, len(directorApplications))
	actions, err := s.reconciler.Do(applications)
	if err != nil {
		return nil, err
	}

	for i, action := range actions {
		results[i] = s.apply(action)
	}

	return results, nil
}

func (s *service) apply(action sync.ApplicationAction) Result {

	app := action.Application
	operation := action.Operation
	serviceActions := action.ServiceActions

	var err apperrors.AppError

	switch action.Operation {
	case sync.Create:
		err = s.applyCreateOperation(app, serviceActions)
	case sync.Delete:
		err = s.applyDeleteOperation(app, serviceActions)
	case sync.Update:
		err = s.applyUpdateOperation(app, serviceActions)
	}

	return newResult(app, operation, err)
}

func (s *service) applyCreateOperation(application v1alpha1.Application, serviceActionActions []sync.ServiceAction) apperrors.AppError {

	var err apperrors.AppError

	_, e := s.applicationRepository.Create(&application)
	if e != nil {
		err = appendError(err, apperrors.Internal("Failed to create Application: %s", e))
	}

	for _, apiDefinition := range application.Spec.Services {
		e := s.resourcesService.CreateApiResources(application, apiDefinition)
		if err != nil {
			err = appendError(err, e)
		}

		e = s.resourcesService.CreateSecrets(application, apiDefinition)
		if err != nil {
			err = appendError(err, e)
		}
	}

	return err
}

func (s *service) applyUpdateOperation(application v1alpha1.Application, apiActions []sync.ServiceAction) apperrors.AppError {
	var err apperrors.AppError

	s.applyApiAndEventActions(application, apiActions)

	{
		_, e := s.applicationRepository.Update(&application)
		if e != nil {
			err = appendError(err, apperrors.Internal("Failed to update application: %s", e))
		}
	}

	return err
}

func (s *service) applyDeleteOperation(application v1alpha1.Application, apiActions []sync.ServiceAction) apperrors.AppError {
	var err apperrors.AppError

	_, _, e := s.applyApiAndEventActions(application, apiActions)
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

func (s *service) applyApiAndEventActions(application v1alpha1.Application, apiActions []sync.ServiceAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {
	err = s.applyApiResources(application, apiActions)

	credentials, params, e := s.applyApiSecrets(application, apiActions)
	if e != nil {
		err = appendError(err, e)
	}

	return credentials, params, err
}

func (s *service) applyApiResources(application v1alpha1.Application, apiActions []sync.ServiceAction) apperrors.AppError {

	var err apperrors.AppError
	for _, apiAction := range apiActions {
		switch apiAction.Operation {
		case sync.Create:
			e := s.resourcesService.CreateApiResources(application, apiAction.Service)
			err = appendError(err, e)
		case sync.Update:
			e := s.resourcesService.UpdateApiResources(application, apiAction.Service)
			err = appendError(err, e)
		case sync.Delete:
			e := s.resourcesService.DeleteApiResources(application, apiAction.Service)
			err = appendError(err, e)
		}
	}

	return err
}

func (s *service) applyApiSecrets(application v1alpha1.Application, APIActions []sync.ServiceAction) (credentials ApiIDToSecretNameMap, params ApiIDToSecretNameMap, err apperrors.AppError) {

	credentials = make(map[string]string)
	params = make(map[string]string)

	for _, apiAction := range APIActions {
		switch apiAction.Operation {
		case sync.Create:
			e := s.resourcesService.CreateSecrets(application, apiAction.Service)
			if err != nil {
				err = appendError(err, e)
			}
		case sync.Update:
			e := s.resourcesService.UpdateSecrets(application, apiAction.Service)
			if err != nil {
				err = appendError(err, e)
			}
		case sync.Delete:
			e := s.resourcesService.DeleteSecrets(application, apiAction.Service)
			err = appendError(err, e)
		}
	}

	return credentials, params, err
}

func newResult(application v1alpha1.Application, operation sync.Operation, appError apperrors.AppError) Result {
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
