package minio

import (
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMinioService_Create(t *testing.T) {

	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	apiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	const bucketName = "content"

	t.Run("should create all specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpec).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpec).Return(nil)

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, eventsSpec)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert documentation if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpec).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpec).Return(nil)

		var emptyDocs []byte

		// when
		apperr := service.Put("1111-2222", emptyDocs, apiSpec, eventsSpec)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert api spec if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpec).Return(nil)

		var emptyApiSpec []byte

		// when
		apperr := service.Put("1111-2222", documentation, emptyApiSpec, eventsSpec)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert events spec if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpec).Return(nil)

		var emptyEventsSpec []byte

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, emptyEventsSpec)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating documentation", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, eventsSpec)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating api spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpec).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, eventsSpec)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating events spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentation).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpec).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpec).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, eventsSpec)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when deleting before put", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentation, apiSpec, eventsSpec)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

}

func TestMinioService_Get(t *testing.T) {

	expectedDocumentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	expectedApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	expectedEventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("should get all specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentation, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(expectedApiSpec, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(expectedEventsSpec, nil)

		// when
		documentation, apiSpec, eventsSpec, apperr := service.Get("1111-2222")

		// then
		require.NoError(t, apperr)
		assert.Equal(t, expectedDocumentation, documentation)
		assert.Equal(t, expectedApiSpec, apiSpec)
		assert.Equal(t, expectedEventsSpec, eventsSpec)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when getting documentation", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(nil, apperrors.Internal(""))

		// when
		_, _, _, apperr := service.Get("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when getting api spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentation, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(nil, apperrors.Internal(""))

		// when
		_, _, _, apperr := service.Get("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when getting events spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentation, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(expectedApiSpec, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(nil, apperrors.Internal(""))

		// when
		_, _, _, apperr := service.Get("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

}

func TestMinioService_Remove(t *testing.T) {
	t.Run("should delete all specs", func(t *testing.T) {
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.json").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.json").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(nil)

		// when
		apperr := service.Remove("1111-2222")

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle failure when removing documentation", func(t *testing.T) {
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.json").Return(apperrors.Internal(""))

		// when
		apperr := service.Remove("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle failure when removing apiSpec", func(t *testing.T) {
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.json").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.json").Return(apperrors.Internal(""))

		// when
		apperr := service.Remove("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle failure when removing eventsSpec", func(t *testing.T) {
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.json").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.json").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(apperrors.Internal(""))

		// when
		apperr := service.Remove("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})
}
