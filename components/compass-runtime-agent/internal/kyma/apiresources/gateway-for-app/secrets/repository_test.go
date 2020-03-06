package secrets

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/mocks"
)

func TestRepository_Create(t *testing.T) {

	t.Run("should create secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Create", secret).Return(secret, nil)

		// when
		err := repository.Create("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should fail if unable to create secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Create", secret).Return(nil, errors.New("some error"))

		// when
		err := repository.Create("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return already exists if secret was already created", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Create", secret).Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		// when
		err := repository.Create("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeAlreadyExists, err.Code())
		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_Get(t *testing.T) {
	t.Run("should get given secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Get", "new-secret", metav1.GetOptions{}).Return(secret, nil)

		// when
		data, err := repository.Get("new-secret")

		// then
		assert.NoError(t, err)
		assert.Equal(t, []byte("testValue1"), data["testKey1"])
		assert.Equal(t, []byte("testValue2"), data["testKey2"])

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error in case fetching fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Get", "secret-name", metav1.GetOptions{}).Return(
			nil,
			errors.New("some error"))

		// when
		data, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		assert.Equal(t, []byte(nil), data["testKey1"])
		assert.Equal(t, []byte(nil), data["testKey2"])

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return not found if secret does not exist", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Get", "secret-name", metav1.GetOptions{}).Return(
			nil,
			k8serrors.NewNotFound(schema.GroupResource{},
				""))

		// when
		data, err := repository.Get("secret-name")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
		assert.NotEmpty(t, err.Error())

		assert.Equal(t, []byte(nil), data["testKey1"])
		assert.Equal(t, []byte(nil), data["testKey2"])

		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_Delete(t *testing.T) {

	t.Run("should delete given secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Delete", "test-secret", &metav1.DeleteOptions{}).Return(
			nil)

		// when
		err := repository.Delete("test-secret")

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)

	})

	t.Run("should return error if deletion fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Delete", "test-secret", &metav1.DeleteOptions{}).Return(
			errors.New("some error"))

		// when
		err := repository.Delete("test-secret")

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should not return error if secret does not exist", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secretsManagerMock.On("Delete", "test-secret", &metav1.DeleteOptions{}).Return(
			k8serrors.NewNotFound(schema.GroupResource{}, ""))

		// when
		err := repository.Delete("test-secret")

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_Upsert(t *testing.T) {

	t.Run("should update secret if it exists", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Update", secret).Return(
			secret, nil)

		// when
		err := repository.Upsert("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should create secret if it does not exist", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Update", secret).Return(
			nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		secretsManagerMock.On("Create", secret).Return(secret, nil)

		// when
		err := repository.Upsert("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error if update fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Update", secret).Return(nil, errors.New("some error"))

		// when
		err := repository.Upsert("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		secretsManagerMock.AssertNotCalled(t, "Create", mock.AnythingOfType("*v1.Secret"))
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error if create fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		repository := NewRepository(secretsManagerMock)

		secret := makeSecret("new-secret", "secretId", "app", "appUID", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})
		secretsManagerMock.On("Update", secret).Return(
			nil, k8serrors.NewNotFound(schema.GroupResource{}, ""))
		secretsManagerMock.On("Create", secret).Return(secret, errors.New("some error"))

		// when
		err := repository.Upsert("app", "appUID", "new-secret", "secretId", map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2"),
		})

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.NotEmpty(t, err.Error())

		secretsManagerMock.AssertExpectations(t)
	})
}
