// Package istio contains components for managing Istio resources (Deniers, DenyRules, CheckNothings, ...)
package istio

import (
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
)

// Service is responsible for creating Istio resources associated with deniers.
type Service interface {
	// Create creates Istio resources associated with deniers.
	Create(remoteEnvironment, serviceId, resourceName string) apperrors.AppError

	// Upsert updates or creates Istio resources associated with deniers.
	Upsert(remoteEnvironment, serviceId, resourceName string) apperrors.AppError

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
func (s *service) Create(remoteEnvironment, serviceId, resourceName string) apperrors.AppError {
	err := s.repository.CreateDenier(remoteEnvironment, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.CreateCheckNothing(remoteEnvironment, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.CreateRule(remoteEnvironment, serviceId, resourceName)
	if err != nil {
		return err
	}

	return nil
}

// Upsert updates or creates Istio resources associated with deniers.
func (s *service) Upsert(remoteEnvironment, serviceId, resourceName string) apperrors.AppError {
	err := s.repository.UpsertDenier(remoteEnvironment, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.UpsertCheckNothing(remoteEnvironment, serviceId, resourceName)
	if err != nil {
		return err
	}

	err = s.repository.UpsertRule(remoteEnvironment, serviceId, resourceName)
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
