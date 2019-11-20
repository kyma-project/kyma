// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository contains operations for managing client credentials
type Repository interface {
	Get(name string) (DexSecret, error)
}

type repository struct {
	secretsManager Manager
}

// Manager contains operations for managing k8s secrets
type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager) Repository {
	return &repository{
		secretsManager: secretsManager,
	}
}

func (r *repository) Get(name string) (DexSecret, error) {
	secret, err := r.secretsManager.Get(name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return DexSecret{}, fmt.Errorf("secret %s not found", name)
		}
		return DexSecret{}, fmt.Errorf("getting %s secret failed: %v", name, err)
	}

	return DexSecret{
		UserEmail:    string(secret.Data[UserEmailKey]),
		UserPassword: string(secret.Data[UserPasswordKey]),
	}, nil
}
