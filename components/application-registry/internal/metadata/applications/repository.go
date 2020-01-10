// Package applications contains components for accessing/modifying Application CRD
package applications

import (
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

// Manager contains operations for managing Application CRD
type Manager interface {
	Update(application *v1alpha1.Application) (*v1alpha1.Application, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

type repository struct {
	reManager Manager
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
type ServiceRepository interface {
	Create(application string, service Service) apperrors.AppError
	Get(application, id string) (Service, apperrors.AppError)
	GetAll(application string) ([]Service, apperrors.AppError)
	Update(application string, service Service) apperrors.AppError
	Delete(application, id string) apperrors.AppError
}

// NewServiceRepository creates a new ApplicationServiceRepository
func NewServiceRepository(reManager Manager) ServiceRepository {
	return &repository{reManager: reManager}
}

// Create adds a new Service in Application
func (r *repository) Create(application string, service Service) apperrors.AppError {
	app, err := r.getApplication(application)
	if err != nil {
		return err
	}

	err = ensureServiceNotExists(service.ID, app)
	if err != nil {
		return err
	}

	app.Spec.Services = append(app.Spec.Services, convertToK8sType(service))

	e := r.updateApplication(app)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Creating service failed, %s", e.Error()))
	}

	return nil
}

// Get reads Service from Application by service id
func (r *repository) Get(application, id string) (Service, apperrors.AppError) {
	re, err := r.getApplication(application)
	if err != nil {
		return Service{}, err
	}

	for _, service := range re.Spec.Services {
		if service.ID == id {
			return convertFromK8sType(service)
		}
	}

	return Service{}, apperrors.NotFound(fmt.Sprintf("Service with ID %s not found", id))
}

// GetAll gets slice of services defined in Application
func (r *repository) GetAll(application string) ([]Service, apperrors.AppError) {
	re, err := r.getApplication(application)
	if err != nil {
		return nil, err
	}

	services := make([]Service, len(re.Spec.Services))
	for i, service := range re.Spec.Services {
		s, err := convertFromK8sType(service)
		if err != nil {
			return nil, err
		}
		services[i] = s
	}

	return services, nil
}

// Update updates a given service defined in Application
func (r *repository) Update(application string, service Service) apperrors.AppError {
	re, err := r.getApplication(application)
	if err != nil {
		return err
	}

	err = ensureServiceExists(service.ID, re)
	if err != nil {
		return err
	}

	replaceService(service.ID, re, convertToK8sType(service))

	e := r.updateApplication(re)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Updating service failed, %s", e.Error()))
	}

	return nil
}

// Delete deletes a given service defined in Application
func (r *repository) Delete(application, id string) apperrors.AppError {
	re, err := r.getApplication(application)
	if err != nil {
		return err
	}

	if !serviceExists(id, re) {
		return nil
	}

	removeService(id, re)

	e := r.updateApplication(re)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Deleting service failed, %s", e.Error()))
	}

	return nil
}

func (r *repository) getApplication(application string) (*v1alpha1.Application, apperrors.AppError) {
	re, err := r.reManager.Get(application, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application %s not found", application)
			return nil, apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Application %s failed, %s", application, err.Error())
		return nil, apperrors.Internal(message)
	}

	return re, nil
}

func (r *repository) updateApplication(re *v1alpha1.Application) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, e := r.reManager.Update(re)
		return e
	})
}
