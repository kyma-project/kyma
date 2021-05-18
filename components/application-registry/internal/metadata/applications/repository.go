// Package applications contains components for accessing/modifying Application CRD
package applications

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	specAPIType                   = "API"
	specEventsType                = "Events"
	CredentialsOAuthType          = "OAuth"
	CredentialsBasicType          = "Basic"
	CredentialsCertificateGenType = "CertificateGen"
)

// AppManager contains operations for managing Application CRD
//go:generate mockery --name AppManager
type AppManager interface {
	Update(ctx context.Context, application *v1alpha1.Application, options v1.UpdateOptions) (*v1alpha1.Application, error)
	Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

type repository struct {
	appManager AppManager
}

// ServiceAPI stores information needed to call an API
type ServiceAPI struct {
	GatewayURL                  string
	AccessLabel                 string
	TargetUrl                   string
	SpecificationUrl            string
	ApiType                     string
	Credentials                 Credentials
	RequestParametersSecretName string
}

type CSRFInfo struct {
	TokenEndpointURL string
}

type Credentials struct {
	Type              string
	SecretName        string
	AuthenticationUrl string
	CSRFInfo          *CSRFInfo
	Headers           *map[string][]string
	QueryParameters   *map[string][]string
}

// Service represents a service stored in Application RE
type Service struct {
	// Mapped to id in Application CRD
	ID string
	// Mapped to identifier in Application CRD
	Identifier string
	// Mapped to displayName in Application CRD
	DisplayName string
	// Mapped to name in Application CRD
	Name string
	// Mapped to shortDescription in Application CRD
	ShortDescription string
	// Mapped to longDescription in Application CRD
	LongDescription string
	// Mapped to labels in Application CRD
	Labels map[string]string
	// Mapped to providerDisplayName in Application CRD
	ProviderDisplayName string
	// Mapped to tags in Application CRD
	Tags []string
	// Mapped to type property under entries element (type: API)
	API *ServiceAPI
	// Mapped to type property under entries element (type: Events)
	Events bool
}

// ServiceRepository contains operations for managing services stored in Application CRD
//go:generate mockery --name ServiceRepository
type ServiceRepository interface {
	Create(appName string, service Service) apperrors.AppError
	Get(appName, id string) (Service, apperrors.AppError)
	GetAll(appName string) ([]Service, apperrors.AppError)
	Update(appName string, service Service) apperrors.AppError
	Delete(appName, id string) apperrors.AppError
	ServiceExists(appName, serviceName string) (bool, apperrors.AppError)
}

// NewServiceRepository creates a new ApplicationServiceRepository
func NewServiceRepository(appManager AppManager) ServiceRepository {
	return &repository{appManager: appManager}
}

// Create adds a new Service in Application
func (r *repository) Create(appName string, service Service) apperrors.AppError {
	err := r.updateApplicationWithRetries(appName, func(app *v1alpha1.Application) error {
		if err := ensureServiceNotExists(service.ID, app); err != nil {
			return err
		}

		app.Spec.Services = append(app.Spec.Services, convertToK8sType(service))
		return nil
	})
	if err != nil {
		return r.plainErrorToInternalAppError(err, "Creating service failed")
	}

	return nil
}

// Get reads Service from Application by service id
func (r *repository) Get(appName, id string) (Service, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return Service{}, err
	}

	for _, service := range app.Spec.Services {
		if service.ID == id {
			return convertFromK8sType(service)
		}
	}

	return Service{}, apperrors.NotFound(fmt.Sprintf("Service with ID %s not found", id))
}

// GetAll gets slice of services defined in Application
func (r *repository) GetAll(appName string) ([]Service, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return nil, err
	}

	services := make([]Service, len(app.Spec.Services))
	for i, service := range app.Spec.Services {
		s, err := convertFromK8sType(service)
		if err != nil {
			return nil, err
		}
		services[i] = s
	}

	return services, nil
}

// ServiceExists returns true if a service with given name is defined in the Application
func (r* repository) ServiceExists(appName, serviceName string) (bool, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return false, err
	}

	for _, service := range app.Spec.Services {
		if serviceName == service.Name {
			return true, nil
		}
	}

	return false, nil
}

// Update updates a given service defined in Application
func (r *repository) Update(appName string, service Service) apperrors.AppError {
	err := r.updateApplicationWithRetries(appName, func(app *v1alpha1.Application) error {
		if err := ensureServiceExists(service.ID, app); err != nil {
			return err
		}

		replaceService(service.ID, app, convertToK8sType(service))
		return nil
	})
	if err != nil {
		return r.plainErrorToInternalAppError(err, "Updating service failed")
	}

	return nil
}

// Delete deletes a given service defined in Application
func (r *repository) Delete(appName, id string) apperrors.AppError {
	err := r.updateApplicationWithRetries(appName, func(app *v1alpha1.Application) error {
		if !serviceExists(id, app) {
			return nil
		}

		removeService(id, app)
		return nil
	})
	if err != nil {
		return r.plainErrorToInternalAppError(err, "Deleting service failed")
	}

	return nil
}

func (r *repository) getApplication(appName string) (*v1alpha1.Application, apperrors.AppError) {
	app, err := r.appManager.Get(context.Background(), appName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application %s not found", appName)
			return nil, apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Application %s failed, %s", appName, err.Error())
		return nil, apperrors.Internal(message)
	}

	return app, nil
}

func (r *repository) updateApplicationWithRetries(
	appName string,
	modifyApplication func(app *v1alpha1.Application) error) error {

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		app, appErr := r.getApplication(appName)
		if appErr != nil {
			return appErr
		}

		err := modifyApplication(app)
		if err != nil {
			return err
		}

		_, err = r.appManager.Update(context.Background(), app, v1.UpdateOptions{})
		return err
	})
}

func (r *repository) plainErrorToInternalAppError(err error, message string) apperrors.AppError {
	appErr, ok := err.(apperrors.AppError)
	if !ok {
		return apperrors.Internal(fmt.Sprintf("%s: %v", message, err))
	}
	return appErr
}
