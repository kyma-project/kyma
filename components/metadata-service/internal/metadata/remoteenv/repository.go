// Package remoteenv contains components for accessing/modifying Remote Environment CRD
package remoteenv

import (
	"fmt"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

const (
	specAPIType          = "API"
	specEventsType       = "Events"
	CredentialsOAuthType = "OAuth"
	CredentialsBasicType = "Basic"
)

// Manager contains operations for managing Remote Environment CRD
type Manager interface {
	Update(*v1alpha1.RemoteEnvironment) (*v1alpha1.RemoteEnvironment, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error)
}

type repository struct {
	reManager Manager
}

// ServiceAPI stores information needed to call an API
type ServiceAPI struct {
	GatewayURL       string
	AccessLabel      string
	TargetUrl        string
	SpecificationUrl string
	ApiType          string
	Credentials      Credentials
}

type Credentials struct {
	Type              string
	SecretName        string
	AuthenticationUrl string
}

// Service represents a service stored in Remote Environment RE
type Service struct {
	// Mapped to id in Remote Environment CRD
	ID string
	// Mapped to identifier in Remote Environment CRD
	Identifier string
	// Mapped to displayName in Remote Environment CRD
	DisplayName string
	// Mapped to name in Remote Environment CRD
	Name string
	// Mapped to shortDescription in Remote Environment CRD
	ShortDescription string
	// Mapped to longDescription in Remote Environment CRD
	LongDescription string
	// Mapped to labels in Remote Environment CRD
	Labels map[string]string
	// Mapped to providerDisplayName in Remote Environment CRD
	ProviderDisplayName string
	// Mapped to tags in Remote Environment CRD
	Tags []string
	// Mapped to type property under entries element (type: API)
	API *ServiceAPI
	// Mapped to type property under entries element (type: Events)
	Events bool
}

// ServiceRepository contains operations for managing services stored in Remote Environment CRD
type ServiceRepository interface {
	Create(remoteEnvironment string, service Service) apperrors.AppError
	Get(remoteEnvironment, id string) (Service, apperrors.AppError)
	GetAll(remoteEnvironment string) ([]Service, apperrors.AppError)
	Update(remoteEnvironment string, service Service) apperrors.AppError
	Delete(remoteEnvironment, id string) apperrors.AppError
}

// NewServiceRepository creates a new RemoteEnvironmentServiceRepository
func NewServiceRepository(reManager Manager) ServiceRepository {
	return &repository{reManager: reManager}
}

// Create adds a new Service in Remote Environment
func (r *repository) Create(remoteEnvironment string, service Service) apperrors.AppError {
	re, err := r.getRemoteEnvironment(remoteEnvironment)
	if err != nil {
		return err
	}

	err = ensureServiceNotExists(service.ID, re)
	if err != nil {
		return err
	}

	re.Spec.Services = append(re.Spec.Services, convertToK8sType(service))

	e := r.updateRemoteEnvironment(re)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Creating service failed, %s", e.Error()))
	}

	return nil
}

// Get reads Service from Remote Environment by service id
func (r *repository) Get(remoteEnvironment, id string) (Service, apperrors.AppError) {
	re, err := r.getRemoteEnvironment(remoteEnvironment)
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

// GetAll gets slice of services defined in Remote Environment
func (r *repository) GetAll(remoteEnvironment string) ([]Service, apperrors.AppError) {
	re, err := r.getRemoteEnvironment(remoteEnvironment)
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

// Update updates a given service defined in Remote Environment
func (r *repository) Update(remoteEnvironment string, service Service) apperrors.AppError {
	re, err := r.getRemoteEnvironment(remoteEnvironment)
	if err != nil {
		return err
	}

	err = ensureServiceExists(service.ID, re)
	if err != nil {
		return err
	}

	replaceService(service.ID, re, convertToK8sType(service))

	e := r.updateRemoteEnvironment(re)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Updating service failed, %s", e.Error()))
	}

	return nil
}

// Delete deletes a given service defined in Remote Environment
func (r *repository) Delete(remoteEnvironment, id string) apperrors.AppError {
	re, err := r.getRemoteEnvironment(remoteEnvironment)
	if err != nil {
		return err
	}

	if !serviceExists(id, re) {
		return nil
	}

	removeService(id, re)

	e := r.updateRemoteEnvironment(re)
	if e != nil {
		return apperrors.Internal(fmt.Sprintf("Deleting service failed, %s", e.Error()))
	}

	return nil
}

func (r *repository) getRemoteEnvironment(remoteEnvironment string) (*v1alpha1.RemoteEnvironment, apperrors.AppError) {
	re, err := r.reManager.Get(remoteEnvironment, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Remote Environment %s not found", remoteEnvironment)
			return nil, apperrors.NotFound(message)
		}

		message := fmt.Sprintf("Getting Remote Environment %s failed, %s", remoteEnvironment, err.Error())
		return nil, apperrors.Internal(message)
	}

	return re, nil
}

func (r *repository) updateRemoteEnvironment(re *v1alpha1.RemoteEnvironment) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, e := r.reManager.Update(re)
		return e
	})
}
