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

	includeAllSourceKeysFunc := func(k string) bool {
		return true
	}

	t.Run("Should rename secret when source and target specified", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(secret, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", targetSecret, secret).Return(nil)
		secretsRepositoryMock.On("Delete", sourceSecret).Return(nil)

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)

	})

	t.Run("Should copy specified keys from source to target secret ", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(secret, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", targetSecret, secret).Return(nil)
		secretsRepositoryMock.On("Delete", sourceSecret).Return(nil)

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)

	})

	t.Run("Should skip copying when source secret name is emppty", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "", Namespace: ""}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secretsRepositoryMock := &mocks.SecretRepository{}

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should skip copying when source secret name is not-emppty but secret doesn't exist", func(t *testing.T) {

		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "source"))

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get source secret", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(map[string][]byte{}, errors.New("failed to get"))

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to get target secret", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(secret, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(map[string][]byte{}, errors.New("failed to get"))

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to create target secret", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(secret, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", targetSecret, secret).Return(errors.New("failed to upsert"))

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should return error when failed to remove source secret", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		secret := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(secret, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(map[string][]byte{}, k8serrors.NewNotFound(schema.GroupResource{}, "target"))
		secretsRepositoryMock.On("Upsert", targetSecret, secret).Return(nil)
		secretsRepositoryMock.On("Delete", sourceSecret).Return(errors.New("failed to upsert"))

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Error(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})

	t.Run("Should remove source secret when target exists", func(t *testing.T) {
		// given
		sourceSecret := types.NamespacedName{Name: "source", Namespace: namespace}
		targetSecret := types.NamespacedName{Name: "target", Namespace: namespace}

		sourceSecretData := map[string][]byte{"key": []byte("value")}
		targetSecretData := map[string][]byte{"key": []byte("value")}

		secretsRepositoryMock := &mocks.SecretRepository{}
		secretsRepositoryMock.On("Get", sourceSecret).Return(sourceSecretData, nil)
		secretsRepositoryMock.On("Get", targetSecret).Return(targetSecretData, nil)
		secretsRepositoryMock.On("Delete", sourceSecret).Return(nil)

		// when
		migrator := NewMigrator(secretsRepositoryMock, includeAllSourceKeysFunc)
		err := migrator.Do(sourceSecret, targetSecret)

		// then
		assert.Nil(t, err)
		secretsRepositoryMock.AssertExpectations(t)
	})
}
