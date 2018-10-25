// Package remoteenv contains components for accessing/modifying Remote Environment CRD
package remoteenv

import (
	"fmt"

	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	specAPIType    = "API"
	specEventsType = "Events"
)

// Manager contains operations for managing Remote Environment CRD
type Manager interface {
	Get(name string, options v1.GetOptions) (*v1alpha1.RemoteEnvironment, error)
}

type repository struct {
	name      string
	reManager Manager
}

// ServiceAPI stores information needed to call an API
type ServiceAPI struct {
	GatewayURL            string
	AccessLabel           string
	TargetUrl             string
	OauthUrl              string
	CredentialsSecretName string
}

// Service represents a service stored in Remote Environment RE
type Service struct {
	// Mapped to id in Remote Environment CRD
	ID string
	// Mapped to displayName in Remote Environment CRD
	DisplayName string
	// Mapped to longDescription in Remote Environment CRD
	LongDescription string
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
	Get(id string) (Service, apperrors.AppError)
}

// NewServiceRepository creates a new RemoteEnvironmentServiceRepository
func NewServiceRepository(name string, reManager Manager) ServiceRepository {
	return &repository{name: name, reManager: reManager}
}

// Get reads Service from Remote Environment by service id
func (r *repository) Get(id string) (Service, apperrors.AppError) {
	re, err := r.getRemoteEnvironment()
	if err != nil {
		return Service{}, err
	}

	for _, service := range re.Spec.Services {
		if service.ID == id {
			return convertFromK8sType(service)
		}
	}

	message := fmt.Sprintf("Service with ID %s not found", id)
	log.Warn(message)

	return Service{}, apperrors.NotFound(message)
}

func (r *repository) getRemoteEnvironment() (*v1alpha1.RemoteEnvironment, apperrors.AppError) {
	re, err := r.reManager.Get(r.name, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Remote environment: %s not found.", r.name)
			log.Warn(message)
			return nil, apperrors.Internal(message)
		}

		message := fmt.Sprintf("failed to get remote environment '%s' : %s", r.name, err.Error())
		log.Error(message)
		return nil, apperrors.Internal(message)
	}

	return re, nil
}
