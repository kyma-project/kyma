package minio

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/minio/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMinioService_Create(t *testing.T) {

	documentationJSON := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	apiSpecJSON := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpecJSON := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	documentationXML := []byte("<description>Some docs blah blah blah</description>")
	apiSpecXML := []byte("<productsEndpoint>Endpoint /products returns products.</productsEndpoint>")
	eventsSpecXML := []byte("<orderCreated>Published when order is placed.</orderCreated>")

	const bucketName = "content"

	t.Run("should create all json specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpecJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpecJSON).Return(nil)

		// when
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, eventsSpecJSON)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should create all xml specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.xml", documentationXML).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.xml", apiSpecXML).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.xml", eventsSpecXML).Return(nil)

		// when
		apperr := service.Put("1111-2222", documentationXML, apiSpecXML, eventsSpecXML)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert documentation if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpecJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpecJSON).Return(nil)

		var emptyDocs []byte

		// when
		apperr := service.Put("1111-2222", emptyDocs, apiSpecJSON, eventsSpecJSON)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert api spec if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpecJSON).Return(nil)

		var emptyApiSpec []byte

		// when
		apperr := service.Put("1111-2222", documentationJSON, emptyApiSpec, eventsSpecJSON)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should not insert events spec if empty", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpecJSON).Return(nil)

		var emptyEventsSpec []byte

		// when
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, emptyEventsSpec)

		// then
		require.NoError(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating documentation", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, eventsSpecJSON)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating api spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpecJSON).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, eventsSpecJSON)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when creating events spec", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, mock.Anything).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/content.json", documentationJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/apiSpec.json", apiSpecJSON).Return(nil)
		repositoryMock.On("Put", bucketName, "service-class/1111-2222/asyncApiSpec.json", eventsSpecJSON).Return(apperrors.Internal(""))

		// when
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, eventsSpecJSON)

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
		apperr := service.Put("1111-2222", documentationJSON, apiSpecJSON, eventsSpecJSON)

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

}

func TestMinioService_Get(t *testing.T) {

	expectedDocumentationJSON := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	expectedApiSpecJSON := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	expectedEventsSpecJSON := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	expectedDocumentationXML := []byte("<description>Some docs blah blah blah</description>")
	expectedApiSpecXML := []byte("<productsEndpoint>Endpoint /products returns products.</productsEndpoint>")
	expectedEventsSpecXML := []byte("<orderCreated>Published when order is placed.</orderCreated>")

	t.Run("should get all json specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentationJSON, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(expectedApiSpecJSON, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(expectedEventsSpecJSON, nil)

		// when
		documentation, apiSpec, eventsSpec, apperr := service.Get("1111-2222")

		// then
		require.NoError(t, apperr)
		assert.Equal(t, expectedDocumentationJSON, documentation)
		assert.Equal(t, expectedApiSpecJSON, apiSpec)
		assert.Equal(t, expectedEventsSpecJSON, eventsSpec)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should get all xml specs", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(nil, apperrors.Internal(""))
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(nil, apperrors.Internal(""))
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(nil, apperrors.Internal(""))

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.xml").Return(expectedDocumentationXML, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.xml").Return(expectedApiSpecXML, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.xml").Return(expectedEventsSpecXML, nil)

		// when
		documentation, apiSpec, eventsSpec, apperr := service.Get("1111-2222")

		// then
		require.NoError(t, apperr)
		assert.Equal(t, expectedDocumentationXML, documentation)
		assert.Equal(t, expectedApiSpecXML, apiSpec)
		assert.Equal(t, expectedEventsSpecXML, eventsSpec)
		repositoryMock.AssertExpectations(t)
	})

	t.Run("should handle errors when getting documentation", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(nil, apperrors.Internal(""))
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.xml").Return(nil, apperrors.Internal(""))

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

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentationJSON, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(nil, apperrors.Internal(""))
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.xml").Return(nil, apperrors.Internal(""))

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

		repositoryMock.On("Get", bucketName, "service-class/1111-2222/content.json").Return(expectedDocumentationJSON, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/apiSpec.json").Return(expectedApiSpecJSON, nil)
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(nil, apperrors.Internal(""))
		repositoryMock.On("Get", bucketName, "service-class/1111-2222/asyncApiSpec.xml").Return(nil, apperrors.Internal(""))

		// when
		_, _, _, apperr := service.Get("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})

}

func TestMinioService_Remove(t *testing.T) {
	t.Run("should delete all json specs", func(t *testing.T) {
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

	t.Run("should delete all xml specs", func(t *testing.T) {
		repositoryMock := &mocks.Repository{}
		service := NewService(repositoryMock)

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.json").Return(apperrors.Internal(""))
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.json").Return(apperrors.Internal(""))
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/asyncApiSpec.json").Return(apperrors.Internal(""))

		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.xml").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.xml").Return(nil)
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/asyncApiSpec.xml").Return(nil)

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
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/content.xml").Return(apperrors.Internal(""))

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
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/apiSpec.xml").Return(apperrors.Internal(""))

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
		repositoryMock.On("Remove", bucketName, "service-class/1111-2222/asyncApiSpec.xml").Return(apperrors.Internal(""))

		// when
		apperr := service.Remove("1111-2222")

		// then
		require.Error(t, apperr)
		repositoryMock.AssertExpectations(t)
	})
}
