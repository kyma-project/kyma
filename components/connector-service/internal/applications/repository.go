// Package applications contains components for accessing/modifying Application CRD
package applications

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager contains operations for managing Application CR
type Manager interface {
	Create(*v1alpha1.Application) (*v1alpha1.Application, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

type repository struct {
	appManager Manager
}

// Repository contains operations for managing Applications CR
type Repository interface {
	Create(application *v1alpha1.Application) apperrors.AppError
	Get(name string) (*v1alpha1.Application, apperrors.AppError)
}

// NewApplicationRepository creates a new Application Repository
func NewApplicationRepository(appManager Manager) Repository {
	return &repository{
		appManager: appManager,
	}
}

// Create creates new Application
func (r *repository) Create(application *v1alpha1.Application) apperrors.AppError {
	_, err := r.appManager.Create(application)
	if err != nil {
		return apperrors.Internal(fmt.Sprintf("Creating application %s failed, %s", application.Name, err.Error()))
	}

	return nil
}

// Get reads Application with name
func (r *repository) Get(name string) (*v1alpha1.Application, apperrors.AppError) {
	app, err := r.appManager.Get(name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, apperrors.NotFound(fmt.Sprintf("Application %s not found, %s", name, err.Error()))
		}

		return nil, apperrors.Internal(fmt.Sprintf("Getting application %s failed, %s", name, err.Error()))
	}

	return app, nil
}
