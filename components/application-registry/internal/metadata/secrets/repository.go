// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Repository contains operations for managing client credentials
type Repository interface {
	Create(application, name, serviceID string, data strategy.SecretData) apperrors.AppError
	Get(application, name string) (strategy.SecretData, apperrors.AppError)
	Delete(name string) apperrors.AppError
	Upsert(application, name, secretID string, data strategy.SecretData) apperrors.AppError
}

type repository struct {
	secretsManager Manager
}

// Manager contains operations for managing k8s secrets
type Manager interface {
	Create(secret *v1.Secret) (*v1.Secret, error)
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Update(secret *v1.Secret) (*v1.Secret, error)
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager Manager) Repository {
	return &repository{
		secretsManager: secretsManager,
	}
}

// Create adds a new secret with one entry containing specified clientId and clientSecret
func (r *repository) Create(application, name, serviceID string, data strategy.SecretData) apperrors.AppError {
	secret := makeSecret(name, serviceID, application, data)
	return r.create(application, secret, name)
}

func (r *repository) Get(application, name string) (data strategy.SecretData, error apperrors.AppError) {
	secret, err := r.secretsManager.Get(name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return strategy.SecretData{}, apperrors.NotFound("Secret %s not found", name)
		}
		return strategy.SecretData{}, apperrors.Internal("Getting %s secret failed, %s", name, err.Error())
	}

	return secret.Data, nil
}

func (r *repository) Delete(name string) apperrors.AppError {
	err := r.secretsManager.Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s secret failed, %s", name, err.Error())
	}
	return nil
}

func (r *repository) Upsert(application, name, serviceID string, data strategy.SecretData) apperrors.AppError {
	secret := makeSecret(name, serviceID, application, data)

	_, err := r.secretsManager.Update(secret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return r.create(application, secret, name)
		}
		return apperrors.Internal("Updating %s secret failed, %s", name, err.Error())
	}
	return nil
}

func (r *repository) create(application string, secret *v1.Secret, name string) apperrors.AppError {
	_, err := r.secretsManager.Create(secret)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return apperrors.AlreadyExists("Secret %s already exists", name)
		}
		return apperrors.Internal("Creating %s secret failed, %s", name, err.Error())
	}
	return nil
}

func makeSecret(name, serviceID, application string, data strategy.SecretData) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelServiceId:   serviceID,
			},
		},
		Data: data,
	}
}
