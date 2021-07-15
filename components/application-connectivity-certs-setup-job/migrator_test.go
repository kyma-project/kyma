package main

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/mocks"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestMigrator(t *testing.T) {

	t.Run("Should rename secret when source and target specified", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(secret, nil)
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: targetSecretName, Namespace: namespace}).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", types.NamespacedName{Name: targetSecretName, Namespace: namespace}, secret).Return(nil)
		secretsRepositoryMock.On("Delete", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(nil)

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)

	})

	t.Run("Should skip copying when source secret name is emppty", func(t *testing.T) {
		// given
		sourceSecretName := ""
		targetSecretName := "target"

		secretsRepositoryMock := &mocks.SecretRepository{}

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should skip copying when source secret name is not-emppty but secret doesn't exist", func(t *testing.T) {
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "source"))

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get source secret", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(map[string][]byte{}, errors.New("failed to get"))

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get target secret", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(secret, nil)
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: targetSecretName, Namespace: namespace}).Return(map[string][]byte{}, errors.New("failed to get"))

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to create target secret", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(secret, nil)
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: targetSecretName, Namespace: namespace}).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", types.NamespacedName{Name: targetSecretName, Namespace: namespace}, secret).Return(errors.New("failed to upsert"))

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to remove source secret", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(secret, nil)
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: targetSecretName, Namespace: namespace}).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", types.NamespacedName{Name: targetSecretName, Namespace: namespace}, secret).Return(nil)
		secretsRepositoryMock.On("Delete", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(errors.New("failed to upsert"))

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should remove source secret when target exists", func(t *testing.T) {
		// given
		sourceSecretName := "source"
		targetSecretName := "target"
		namespace := "test"

		sourceSecret := map[string][]byte{"key": []byte("value")}
		targetSecret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(sourceSecret, nil)
		secretsRepositoryMock.On("Get", types.NamespacedName{Name: targetSecretName, Namespace: namespace}).Return(targetSecret, nil)
		secretsRepositoryMock.On("Delete", types.NamespacedName{Name: sourceSecretName, Namespace: namespace}).Return(nil)

		// when
		migrator := NewMigrator(secretsRepositoryMock)
		err := migrator.Do(sourceSecretName, targetSecretName, namespace)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})
}
