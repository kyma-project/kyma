package revocation

import (
	"context"
	"errors"
	"testing"

	k8sclientMocks "github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

var testContext = context.Background()

func TestRevocationListRepository(t *testing.T) {

	configMapName := "revokedCertificates"

	t.Run("should return false if value is not present", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}
		configMapName := "revokedCertificates"

		configListManagerMock.On("Get", testContext, configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		isPresent, err := repository.Contains(testContext, someHash)
		require.NoError(t, err)

		// then
		assert.Equal(t, isPresent, false)
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should insert value to the list", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", testContext, configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", testContext, &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}, mock.AnythingOfType("v1.UpdateOptions")).Return(&v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}, nil)

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(context.Background(), someHash)
		require.NoError(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to get config map", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", testContext, configMapName, mock.AnythingOfType("v1.GetOptions")).Return(nil, errors.New("some error"))

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(testContext, someHash)
		require.Error(t, err)

		_, err = repository.Contains(testContext, someHash)
		require.Error(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to update config map", func(t *testing.T) {
		// given
		someHash := "someHash"
		configListManagerMock := &k8sclientMocks.Manager{}

		configListManagerMock.On("Get", testContext, configMapName, mock.AnythingOfType("v1.GetOptions")).Return(
			&v1.ConfigMap{
				Data: nil,
			}, nil)

		configListManagerMock.On("Update", testContext, &v1.ConfigMap{
			Data: map[string]string{
				someHash: someHash,
			}}, mock.AnythingOfType("v1.UpdateOptions")).Return(nil, errors.New("some error"))

		repository := NewRepository(configListManagerMock, configMapName)

		// when
		err := repository.Insert(testContext, someHash)
		require.Error(t, err)

		// then
		configListManagerMock.AssertExpectations(t)
	})
}
