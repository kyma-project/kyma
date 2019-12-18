package helpers

import (
"fmt"

v1 "k8s.io/api/core/v1"
k8serrors "k8s.io/apimachinery/pkg/api/errors"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretRepository contains operations for managing client credentials
type SecretRepository interface {
	Get(name string) (DexSecret, error)
}

type secretRepository struct {
	secretsManager SecretManager
}

const (
	UserEmailKey    = "email"
	UserPasswordKey = "password"
)

type DexSecret struct {
	UserEmail    string
	UserPassword string
}

// SecretManager contains operations for managing k8s secrets
type SecretManager interface {
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

// NewSecretRepository creates a new secrets secretRepository
func NewSecretRepository(secretsManager SecretManager) SecretRepository {
	return &secretRepository{
		secretsManager: secretsManager,
	}
}

func (r *secretRepository) Get(name string) (DexSecret, error) {
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
