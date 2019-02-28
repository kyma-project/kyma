package secrets

import (
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dataKey    = "dataKey"
	secretName = "secret-name"
)

var (
	secretData = map[string][]byte{
		"testKey2": []byte("testValue2"),
		"testKey1": []byte("testValue1"),
	}
)

func TestRepository_Get(t *testing.T) {
	t.Run("should get given secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, map[string][]byte{dataKey: []byte("data")})

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(secret, nil)

		repository := NewRepository(secretsManagerMock)

		// when
		secrets, err := repository.Get(secretName)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, secrets[dataKey])

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error in case fetching fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(
			nil,
			errors.New("some error"))

		repository := NewRepository(secretsManagerMock)

		// when
		secretData, err := repository.Get(secretName)

		// then
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
		assert.Nil(t, secretData)

		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_Override(t *testing.T) {

	t.Run("should create secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Create", secret).Return(secret, nil)

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithReplace(secretName, secretData)

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if unable to create secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Create", secret).Return(nil, errors.New("some error"))

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithReplace(secretName, secretData)

		// then
		require.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should override secret if already exist", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Create", secret).Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, "error")).Once()
		secretsManagerMock.On("Create", secret).Return(nil, nil).Once()
		secretsManagerMock.On("Delete", secretName, &metav1.DeleteOptions{}).Return(nil)

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithReplace(secretName, secretData)

		// then
		require.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error if failed to delete secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Create", secret).Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Delete", secretName, &metav1.DeleteOptions{}).Return(errors.New("error"))

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithReplace(secretName, secretData)

		// then
		require.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error if failed to create secret after deleting", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Create", secret).Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, "error")).Once()
		secretsManagerMock.On("Create", secret).Return(nil, errors.New("error")).Once()
		secretsManagerMock.On("Delete", secretName, &metav1.DeleteOptions{}).Return(nil)

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithReplace(secretName, secretData)

		// then
		require.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_UpsertData(t *testing.T) {

	t.Run("should update secret data if it exists", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		additionalSecretData := map[string][]byte{
			"testKey2": []byte("testValue2Modified"),
			"testKey3": []byte("testValue3"),
		}

		updatedSecret := makeSecret(secretName, map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2Modified"),
			"testKey3": []byte("testValue3"),
		})

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(secret, nil)
		secretsManagerMock.On("Update", updatedSecret).Return(secret, nil)

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithMerge(secretName, additionalSecretData)

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should create new secret if it does not exists", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", secret).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Create", secret).Return(secret, nil)

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithMerge(secretName, secretData)

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(nil, errors.New("error"))

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithMerge(secretName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to update secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", secret).Return(nil, errors.New("error"))

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithMerge(secretName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to create secret", func(t *testing.T) {
		// given
		secret := makeSecret(secretName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", secret).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Create", secret).Return(nil, errors.New("error"))

		repository := NewRepository(secretsManagerMock)

		// when
		err := repository.UpsertWithMerge(secretName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

}
