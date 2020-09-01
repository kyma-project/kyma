// Package istio contains components for managing Istio AuthorizationPolicies
package istio

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"k8s.io/apimachinery/pkg/types"
)

// Service is responsible for creating Istio AuthorizationPolicy.
type Service interface {
	// Create creates Istio AuthorizationPolicy.
	Create(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError

	// Upsert updates or creates Istio AuthorizationPolicy.
	Upsert(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError

	// Delete removes Istio AuthorizationPolicy.
	Delete(resourceName string) apperrors.AppError
}

type service struct {
	repository Repository
}

// NewService creates a new Service.
func NewService(repository Repository) Service {
	return &service{repository: repository}
}

// Create creates Istio AuthorizationPolicy.
func (s *service) Create(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError {
	return s.repository.CreateAuthorizationPolicy(application, appUID, serviceId, resourceName)
}

// Upsert updates or creates Istio AuthorizationPolicy.
func (s *service) Upsert(application string, appUID types.UID, serviceId, resourceName string) apperrors.AppError {
	return s.repository.UpsertAuthorizationPolicy(application, appUID, serviceId, resourceName)
}

// Delete removes Istio AuthorizationPolicy.
func (s *service) Delete(resourceName string) apperrors.AppError {
	return s.repository.DeleteAuthorizationPolicy(resourceName)
}
