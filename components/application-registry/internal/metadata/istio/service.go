// Package istio contains components for managing Istio resources (Deniers, DenyRules, CheckNothings, ...)
package istio

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"k8s.io/apimachinery/pkg/types"
)

// Service is responsible for creating Istio resources associated with deniers.
type Service interface {
	// Create creates Istio resources associated with deniers.
	Create(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError

	// Upsert updates or creates Istio resources associated with deniers.
	Upsert(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError

	// Delete removes Istio resources associated with deniers.
	Delete(resourceName string) apperrors.AppError
}

type service struct {
	repository Repository
}

// NewService creates a new Service.
func NewService(repository Repository) Service {
	return &service{repository: repository}
}

// Create creates Istio resources associated with deniers.
func (s *service) Create(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError {
	err := s.repository.CreateDenier(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.CreateCheckNothing(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.CreateRule(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	return nil
}

// Upsert updates or creates Istio resources associated with deniers.
func (s *service) Upsert(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError {
	err := s.repository.UpsertDenier(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.UpsertCheckNothing(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.UpsertRule(application, appUID, serviceId, resourceName)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes Istio resources associated with deniers.
func (s *service) Delete(resourceName string) apperrors.AppError {
	err := s.repository.DeleteDenier(resourceName)
	if err != nil {
		return err
	}

	err = s.repository.DeleteCheckNothing(resourceName)
	if err != nil {
		return err
	}

	err = s.repository.DeleteRule(resourceName)
	if err != nil {
		return err
	}

	return nil
}
