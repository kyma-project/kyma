// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "github.com/sirupsen/logrus"
)

// Repository contains operations for managing client credentials
type Repository interface {
	Get(name string) (map[string][]byte, apperrors.AppError)
}

type repository struct {
	secretsManager    Manager
	remoteEnvironment string
}

// Manager contains operations for managing k8s secrets
type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager, remoteEnvironment string) Repository {
	return &repository{
		secretsManager:    secretsManager,
		remoteEnvironment: remoteEnvironment,
	}
}

func (r *repository) Get(name string) (map[string][]byte, apperrors.AppError) {
	secret, err := r.secretsManager.Get(name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to read secret '%s': %s", name, err.Error())
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("secret '%s' not found", name)
		}
		return nil, apperrors.Internal("failed to get '%s' secret, %s", name, err)
	}

	return secret.Data, nil
}
