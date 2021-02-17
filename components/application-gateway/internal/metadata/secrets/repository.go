// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"context"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository contains operations for managing client credentials
//go:generate mockery --name=Repository
type Repository interface {
	Get(name string) (map[string][]byte, apperrors.AppError)
}

type repository struct {
	secretsManager Manager
	application    string
}

// Manager contains operations for managing k8s secrets
//go:generate mockery --name=Manager
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager, application string) Repository {
	return &repository{
		secretsManager: secretsManager,
		application:    application,
	}
}

func (r *repository) Get(name string) (map[string][]byte, apperrors.AppError) {
	secret, err := r.secretsManager.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to read secret '%s': %s", name, err.Error())
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("secret '%s' not found", name)
		}
		return nil, apperrors.Internal("failed to get '%s' secret, %s", name, err)
	}

	return secret.Data, nil
}
