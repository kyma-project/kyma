package kyma

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Service
type Service interface {
	Apply(applications []model.Application) ([]Result, apperrors.AppError)
}

func NewSynchronizationService() Service {
	return &service{}
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

func NewService(applicationRepository applications.Manager, converter applications.Converter, resourcesService apiresources.Service) Service {
	return &service{
		applicationRepository: applicationRepository,
		converter:             converter,
		resourcesService:      resourcesService,
	}
}

func (s *service) Apply(directorApplications []model.Application) ([]Result, apperrors.AppError) {
	logrus.Info("Application passed to Sync service: ", len(directorApplications))

	currentApplications, err := s.getCurrentApplications()
	if err != nil {
		return nil, err
	}

	return s.apply(currentApplications, directorApplications), nil
}

func (s *service) getCurrentApplications() ([]v1alpha1.Application, apperrors.AppError) {
	applications, err := s.applicationRepository.List(v1.ListOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to get application list: %s", err)
	}

	return applications.Items, nil
}

func (s *service) apply(currentApplications []v1alpha1.Application, directorApplications []model.Application) []Result {
	results := make([]Result, 0)

	created := s.createApplications(currentApplications, directorApplications)
	deleted := s.deleteApplications(currentApplications, directorApplications)
	updated := s.updateApplications(currentApplications, directorApplications)

	results = append(results, created...)
	results = append(results, deleted...)
	results = append(results, updated...)

	return results
}

func (s *service) createApplications(currentApplications []v1alpha1.Application, directorApplications []model.Application) []Result {

	results := make([]Result, 0)

	for _, directorApplication := range directorApplications {
		if !applications.ApplicationExists(directorApplication.ID, currentApplications) {
			r := s.createApplication(directorApplication, s.converter.Do(directorApplication))
			results = append(results, r)
		}
	}

	return results
}

func (s *service) createApplication(directorApplication model.Application, runtimeApplication v1alpha1.Application) Result {

	err := s.createAPIResources(directorApplication, runtimeApplication)

	_, e := s.applicationRepository.Create(&runtimeApplication)
	if e != nil {
		err = appendError(err, apperrors.Internal("Failed to create application: %s", e))
	}

	return newResult(runtimeApplication, Create, nil)
}

func (s *service) createAPIResources(directorApplication model.Application, runtimeApplication v1alpha1.Application) apperrors.AppError {
	var err apperrors.AppError

	for _, apiDefinition := range directorApplication.APIs {
		spec := getSpec(apiDefinition.APISpec)
		service := applications.GetService(apiDefinition.ID, runtimeApplication)

		e := s.resourcesService.CreateApiResources(runtimeApplication, service, spec)
		if e != nil {
			appendError(err, e)
		}
	}

	for _, eventApiDefinition := range directorApplication.EventAPIs {
		spec := getEventSpec(eventApiDefinition.EventAPISpec)
		service := applications.GetService(eventApiDefinition.ID, runtimeApplication)

		e := s.resourcesService.CreateApiResources(runtimeApplication, service, spec)
		if e != nil {
			appendError(err, e)
		}
	}

	return err
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

func (s *service) deleteApplications(currentApplications []v1alpha1.Application, directorApplications []model.Application) []Result {

	return nil
}

func (s *service) updateApplications(currentApplications []v1alpha1.Application, directorApplications []model.Application) []Result {

	return nil
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
