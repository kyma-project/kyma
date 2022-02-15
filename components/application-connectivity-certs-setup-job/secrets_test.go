package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"

	"github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/mocks"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	dataKey    = "dataKey"
	secretName = "secret-name"
	namespace  = "kyma-integration"
)

var (
	testContext = context.Background()

	namespacedName = types.NamespacedName{
		Name:      secretName,
		Namespace: namespace,
	}

	secretData = map[string][]byte{
		"testKey2": []byte("testValue2"),
		"testKey1": []byte("testValue1"),
	}
)

func TestRepository_Get(t *testing.T) {
	t.Run("should get given secret", func(t *testing.T) {
		// given
		secret := makeSecret(namespacedName, map[string][]byte{dataKey: []byte("data")})

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(secret, nil)

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		secrets, err := repository.Get(namespacedName)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, secrets[dataKey])

		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return an error in case fetching fails", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(
			nil,
			errors.New("some error"))

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		secretData, err := repository.Get(namespacedName)

		// then
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
		assert.Nil(t, secretData)

		secretsManagerMock.AssertExpectations(t)
	})
}

func TestRepository_Upsert(t *testing.T) {

	t.Run("should update secret data if it exists", func(t *testing.T) {
		// given
		secret := makeSecret(namespacedName, secretData)

		additionalSecretData := map[string][]byte{
			"testKey2": []byte("testValue2Modified"),
			"testKey3": []byte("testValue3"),
		}

		updatedSecret := makeSecret(namespacedName, map[string][]byte{
			"testKey1": []byte("testValue1"),
			"testKey2": []byte("testValue2Modified"),
			"testKey3": []byte("testValue3"),
		})

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(secret, nil)
		secretsManagerMock.On("Update", testContext, updatedSecret, metav1.UpdateOptions{}).Return(secret, nil)

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		err := repository.Upsert(namespacedName, additionalSecretData)

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should create new secret if it does not exists", func(t *testing.T) {
		// given
		secret := makeSecret(namespacedName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", testContext, secret, metav1.UpdateOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Create", testContext, secret, metav1.CreateOptions{}).Return(secret, nil)

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		err := repository.Upsert(namespacedName, secretData)

		// then
		assert.NoError(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(nil, errors.New("error"))

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		err := repository.Upsert(namespacedName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to update secret", func(t *testing.T) {
		// given
		secret := makeSecret(namespacedName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", testContext, secret, metav1.UpdateOptions{}).Return(nil, errors.New("error"))

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		err := repository.Upsert(namespacedName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to create secret", func(t *testing.T) {
		// given
		secret := makeSecret(namespacedName, secretData)

		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Update", testContext, secret, metav1.UpdateOptions{}).Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		secretsManagerMock.On("Create", testContext, secret, metav1.CreateOptions{}).Return(nil, errors.New("error"))

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		err := repository.Upsert(namespacedName, secretData)

		// then
		assert.Error(t, err)
		secretsManagerMock.AssertExpectations(t)
	})

}

func TestRepository_ValuesProvided(t *testing.T) {

	testCases := []struct {
		description    string
		secret         *v1.Secret
		error          error
		searchedKeys   []string
		valuesProvided bool
	}{
		{
			description: "should return true if values provided",
			secret: makeSecret(namespacedName, map[string][]byte{
				dataKey: []byte("value"),
			}),
			searchedKeys:   []string{dataKey},
			valuesProvided: true,
		},
		{
			description: "should return false if value empty",
			secret: makeSecret(namespacedName, map[string][]byte{
				dataKey: []byte(""),
			}),
			searchedKeys:   []string{dataKey},
			valuesProvided: false,
		},
		{
			description: "should return false if at least one value is empty",
			secret: makeSecret(namespacedName, map[string][]byte{
				dataKey:      []byte("data"),
				"anotherKey": []byte(""),
			}),
			searchedKeys:   []string{dataKey, "anotherKey"},
			valuesProvided: false,
		},
		{
			description:    "should return false if value does not exist",
			secret:         makeSecret(namespacedName, map[string][]byte{}),
			searchedKeys:   []string{dataKey},
			valuesProvided: false,
		},
		{
			description:    "should return false if secret not found",
			secret:         nil,
			error:          k8serrors.NewNotFound(schema.GroupResource{}, "error"),
			searchedKeys:   []string{dataKey},
			valuesProvided: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			// given
			secretsManagerMock := &mocks.Manager{}
			secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(test.secret, test.error)

			repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

			// when
			provided, err := repository.ValuesProvided(namespacedName, test.searchedKeys)

			// then
			require.NoError(t, err)
			assert.Equal(t, test.valuesProvided, provided)
		})
	}

	t.Run("should return error when failed to get secret", func(t *testing.T) {
		// given
		secretsManagerMock := &mocks.Manager{}
		secretsManagerMock.On("Get", testContext, secretName, metav1.GetOptions{}).Return(nil, errors.New("error"))

		repository := NewSecretRepository(prepareManagerConstructor(secretsManagerMock))

		// when
		_, err := repository.ValuesProvided(namespacedName, []string{dataKey})

		// then
		require.Error(t, err)
	})

}

func prepareManagerConstructor(manager Manager) ManagerConstructor {
	return func(namespace string) Manager {
		return manager
	}
}
