// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"context"
	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository contains operations for managing client credentials
//
//go:generate mockery --name=Repository
type Repository interface {
	Get(name string) (map[string][]byte, apperrors.AppError)
}

type repository struct {
	secretsManager Manager
	application    string
}

// Manager contains operations for managing k8s secrets
//
//go:generate mockery --name=Manager
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager) Repository {
	return &repository{
		secretsManager: secretsManager,
	}
}

func (r *repository) Get(name string) (map[string][]byte, apperrors.AppError) {
	secret, err := r.secretsManager.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		zap.L().Error("failed to read secret",
			zap.String("secretName", name),
			zap.Error(err))
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("secret '%s' not found", name)
		}
		return nil, apperrors.Internal("failed to get '%s' secret, %s", name, err)
	}

	return secret.Data, nil
}
