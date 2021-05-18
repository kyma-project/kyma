// Package appsecrets contains components for accessing/modifying client secrets
package appsecrets

import (
	"context"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/secrets/strategy"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Repository contains operations for managing client credentials
//go:generate mockery --name Repository
type Repository interface {
	Create(application string, appUID types.UID, name, packageID string, data strategy.SecretData) apperrors.AppError
	Get(name string) (strategy.SecretData, apperrors.AppError)
	Delete(name string) apperrors.AppError
	Upsert(application string, appUID types.UID, name, packageID string, data strategy.SecretData) apperrors.AppError
}

type repository struct {
	secretsManager secrets.Manager
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManager secrets.Manager) Repository {
	return &repository{
		secretsManager: secretsManager,
	}
}

// Create adds a new secret with one entry containing specified clientId and clientSecret
func (r *repository) Create(application string, appUID types.UID, name, packageID string, data strategy.SecretData) apperrors.AppError {
	secret := makeSecret(name, packageID, application, appUID, data)
	return r.create(secret, name)
}

func (r *repository) Get(name string) (data strategy.SecretData, error apperrors.AppError) {
	secret, err := r.secretsManager.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return strategy.SecretData{}, apperrors.NotFound("Secret %s not found", name)
		}
		return strategy.SecretData{}, apperrors.Internal("Getting %s secret failed, %s", name, err.Error())
	}

	return secret.Data, nil
}

func (r *repository) Delete(name string) apperrors.AppError {
	err := r.secretsManager.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return apperrors.Internal("Deleting %s secret failed, %s", name, err.Error())
	}
	return nil
}

func (r *repository) Upsert(application string, appUID types.UID, name, packageID string, data strategy.SecretData) apperrors.AppError {
	secret := makeSecret(name, packageID, application, appUID, data)

	_, err := r.secretsManager.Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return r.create(secret, name)
		}
		return apperrors.Internal("Updating %s secret failed, %s", name, err.Error())
	}
	return nil
}

func (r *repository) create(secret *v1.Secret, name string) apperrors.AppError {
	_, err := r.secretsManager.Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return apperrors.AlreadyExists("Secret %s already exists", name)
		}
		return apperrors.Internal("Creating %s secret failed, %s", name, err.Error())
	}
	return nil
}

func makeSecret(name, packageID, application string, appUID types.UID, data strategy.SecretData) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				k8sconsts.LabelApplication: application,
				k8sconsts.LabelPackageId:   packageID,
			},
			OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication(application, appUID),
		},
		Data: data,
	}
}
