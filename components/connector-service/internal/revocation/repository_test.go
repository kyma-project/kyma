package revocation

import (
	"errors"
	"testing"

	k8sclientMocks "github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestRevocationListRepository(t *testing.T) {

	configMapName := "revokedCertificates"

	t.Run("should return false if value is not present", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}
		configMapName := "revokedCertificates"

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		isPresent, err := repository.Contains(someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, false)
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should insert value to the list", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}).Return(&v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}, nil)

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(someHash)
		require.NoError(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to get config map", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(nil, errors.New("some error"))

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(someHash)
		require.Error(t, err)

		_, err = repository.Contains(someHash)
		require.Error(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to update config map", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}).Return(nil, errors.New("some error"))

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(someHash)
		require.Error(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})
}
