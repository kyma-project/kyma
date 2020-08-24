package applications

import (
	"context"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type manager struct {
	applicationsInterface v1alpha12.ApplicationInterface
}

//go:generate mockery --name=Repository
// Repository contains operations for managing Application CRD
type Repository interface {
	Create(*v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError)
	Update(*v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError)
	Delete(name string, options *metav1.DeleteOptions) apperrors.AppError
	Get(name string, options metav1.GetOptions) (*v1alpha1.Application, apperrors.AppError)
	List(opts metav1.ListOptions) (*v1alpha1.ApplicationList, apperrors.AppError)
}

func NewRepository(applicationsInterface v1alpha12.ApplicationInterface) Repository {
	return manager{
		applicationsInterface: applicationsInterface,
	}
}

func (m manager) Create(application *v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError) {

	app, err := m.applicationsInterface.Create(context.Background(), application, metav1.CreateOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to create application: %s", err)
	}

	return app, nil
}

func (m manager) Update(application *v1alpha1.Application) (*v1alpha1.Application, apperrors.AppError) {
	currentApp, err := m.applicationsInterface.Get(context.Background(), application.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("Failed to update application: %s", err)
		}
	}

	currentApp.Spec.Description = application.Spec.Description
	currentApp.Spec.Labels = application.Spec.Labels
	currentApp.Spec.Services = application.Spec.Services
	currentApp.Spec.CompassMetadata = application.Spec.CompassMetadata

	newApp, err := m.applicationsInterface.Update(context.Background(), currentApp, metav1.UpdateOptions{})
	if err != nil {
		return nil, apperrors.Internal("Failed to update application: %s", err)
	}

	return newApp, nil
}

func (m manager) Delete(name string, options *metav1.DeleteOptions) apperrors.AppError {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	err := m.applicationsInterface.Delete(context.Background(), name, *options)
	if err != nil {
		return apperrors.Internal("Failed to delete application: %s", err)
	}

	return nil
}

func (m manager) Get(name string, options metav1.GetOptions) (*v1alpha1.Application, apperrors.AppError) {
	app, err := m.applicationsInterface.Get(context.Background(), name, options)
	if err != nil {
		return nil, apperrors.Internal("Failed to get application: %s", err)
	}

	return app, nil
}

func (m manager) List(opts metav1.ListOptions) (*v1alpha1.ApplicationList, apperrors.AppError) {
	apps, err := m.applicationsInterface.List(context.Background(), opts)
	if err != nil {
		return nil, apperrors.Internal("Failed to list applications: %s", err)
	}

	return apps, nil
}
