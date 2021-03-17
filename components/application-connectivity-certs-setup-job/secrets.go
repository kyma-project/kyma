package main

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// ManagerConstructor creates Secret Manager for specified namespace
type ManagerConstructor func(namespace string) Manager

// Manager contains operations for managing k8s secrets
type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.Secret, error)
	Create(ctx context.Context, secret *v1.Secret, options metav1.CreateOptions) (*v1.Secret, error)
	Update(ctx context.Context, secret *v1.Secret, options metav1.UpdateOptions) (*v1.Secret, error)
}

// SecretRepository contains operations for managing secrets
type SecretRepository interface {
	Get(name types.NamespacedName) (map[string][]byte, error)
	Upsert(name types.NamespacedName, data map[string][]byte) error
	ValuesProvided(secretName types.NamespacedName, keys []string) (bool, error)
}

type repository struct {
	ctx                       context.Context
	secretsManagerConstructor ManagerConstructor
	application               string
}

// NewRepository creates a new secrets repository
func NewSecretRepository(secretsManagerConstructor ManagerConstructor) SecretRepository {
	return &repository{
		ctx:                       context.Background(),
		secretsManagerConstructor: secretsManagerConstructor,
	}
}

// ValuesProvided specifies if the secret exists with not empty values for provided keys
func (r *repository) ValuesProvided(secretName types.NamespacedName, keys []string) (bool, error) {
	secretManager := r.secretsManagerConstructor(secretName.Namespace)

	secret, err := secretManager.Get(r.ctx, secretName.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}

		return false, errors.Wrap(err, "Failed to get secret")
	}

	for _, key := range keys {
		if len(secret.Data[key]) == 0 {
			return false, nil
		}
	}

	return true, nil
}

// Get returns secret data for specified name
func (r *repository) Get(name types.NamespacedName) (map[string][]byte, error) {
	secretManager := r.secretsManagerConstructor(name.Namespace)

	secret, err := secretManager.Get(r.ctx, name.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}

// Upsert updates secrets data with the provided values. If provided value already exists it will be updated.
// If secret does not exist it will be created
func (r *repository) Upsert(name types.NamespacedName, data map[string][]byte) error {
	existingData, err := r.Get(name)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Wrap(err, "Failed to upsert secret data")
		}

		existingData = map[string][]byte{}
	}

	mergedData := mergeSecretData(existingData, data)
	return r.upsert(name, mergedData)
}

func (r *repository) upsert(name types.NamespacedName, data map[string][]byte) error {
	secretManager := r.secretsManagerConstructor(name.Namespace)

	secret := makeSecret(name, data)

	_, err := secretManager.Update(r.ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = secretManager.Create(r.ctx, secret, metav1.CreateOptions{})
			return err
		}
		return errors.Wrapf(err, fmt.Sprintf("Updating %s secret failed while upserting", name))
	}
	return nil
}

func makeSecret(name types.NamespacedName, data map[string][]byte) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: data,
	}
}

func mergeSecretData(data, newData map[string][]byte) map[string][]byte {
	for key, value := range newData {
		data[key] = value
	}

	return data
}
