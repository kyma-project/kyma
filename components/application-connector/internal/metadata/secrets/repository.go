// Package secrets contains components for accessing/modifying client secrets
package secrets

import (
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/k8sconsts"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClientIDKey     = "clientId"
	ClientSecretKey = "clientSecret"
)

// Repository contains operations for managing client credentials
type Repository interface {
	Create(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError
	Get(remoteEnvironment, name string) (string, string, apperrors.AppError)
	Delete(name string) apperrors.AppError
	Upsert(remoteEnvironment, name, clientID, clientSecret, secretID string) apperrors.AppError
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
func (r *repository) Create(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	secret := makeSecret(name, clientID, clientSecret, serviceID, remoteEnvironment)
	return r.create(remoteEnvironment, secret, name)
}

func (r *repository) Get(remoteEnvironment, name string) (clientId string, clientSecret string, error apperrors.AppError) {
	secret, err := r.secretsManager.Get(name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return "", "", apperrors.NotFound("secret %s not found", name)
		}
		return "", "", apperrors.Internal("failed to get %s secret, %s", name, err)
	}

	return string(secret.Data[ClientIDKey]), string(secret.Data[ClientSecretKey]), nil
}

func (r *repository) Delete(name string) apperrors.AppError {
	err := r.secretsManager.Delete(name, &metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("failed to delete %s secret, %s", name, err)
	}
	return nil
}

func (r *repository) Upsert(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	secret := makeSecret(name, clientID, clientSecret, serviceID, remoteEnvironment)

	_, err := r.secretsManager.Update(secret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return r.create(remoteEnvironment, secret, name)
		}
		return apperrors.Internal("failed to update %s secret, %s", name, err)
	}
	return nil
}

func (r *repository) create(remoteEnvironment string, secret *v1.Secret, name string) apperrors.AppError {
	_, err := r.secretsManager.Create(secret)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return apperrors.AlreadyExists("secret %s already exists.", name)
		}
		return apperrors.Internal("failed to create %s secret, %s", name, err)
	}
	return nil
}

func makeSecret(name, clientID, clientSecret, serviceID, remoteEnvironment string) *v1.Secret {
	secretMap := make(map[string][]byte)
	secretMap[ClientIDKey] = []byte(clientID)
	secretMap[ClientSecretKey] = []byte(clientSecret)

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelRemoteEnvironment: remoteEnvironment,
				k8sconsts.LabelServiceId:         serviceID,
			},
		},
		Data: secretMap,
	}
}
