package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	uploadMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddingToAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should add all specifications to asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			docsTopic := createTestDocsTopic("id1",
				"www.somestorage.com/apiSpec.json",
				"www.somestorage.com/asyncApiSpec.json",
				"www.somestorage.com/content.json", docstopic.StatusNone)

			repositoryMock.On("Update", docsTopic).Return(apperrors.NotFound("Not found."))
			repositoryMock.On("Create", docsTopic).Return(nil)
		}

		{
			specFile := createTestInputFile(openApiSpecFileName, "", jsonApiSpec)
			eventsFile := createTestInputFile(eventsSpecFileName, "", eventsSpec)
			docsFile := createTestInputFile(documentationFileName, "", documentation)

			uploadClientMock.On("Upload", specFile).
				Return(createTestOutputFile(openApiSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", eventsFile).
				Return(createTestOutputFile(eventsSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", docsFile).
				Return(createTestOutputFile(documentationFileName, "www.somestorage.com"), nil)
		}

		// when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData XML specification", func(t *testing.T) {

	})

	t.Run("Should detect OData JSON specification", func(t *testing.T) {

	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func TestGettingFromAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should get specifications from asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		apiTestServer := createTestServer(t, jsonApiSpec)
		defer apiTestServer.Close()

		eventTestServer := createTestServer(t, eventsSpec)
		defer eventTestServer.Close()

		documentationServer := createTestServer(t, documentation)
		defer documentationServer.Close()

		{
			repositoryMock.On("Get", "id1").
				Return(createTestDocsTopic("id1", apiTestServer.URL, eventTestServer.URL, documentationServer.URL, docstopic.StatusSucceeded), nil)
		}

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.NoError(t, err)
		assert.Equal(t, jsonApiSpec, api)
		assert.Equal(t, eventsSpec, events)
		assert.Equal(t, documentation, docs)

		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when DocsTopic status is failed", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			repositoryMock.On("Get", "id1").Return(createTestDocsTopic("id1", "", "", "", docstopic.StatusFailed), nil)
		}

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, api)
		assert.Nil(t, events)
		assert.Nil(t, docs)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("Should fail when DocsTopic status is pending", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			repositoryMock.On("Get", "id1").Return(createTestDocsTopic("id1", "", "", "", docstopic.StatusPending), nil)
		}

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.NoError(t, err)
		assert.Nil(t, api)
		assert.Nil(t, events)
		assert.Nil(t, docs)

		repositoryMock.AssertExpectations(t)
	})
}

func TestUpdatingInAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should update specifications in asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			docsTopic := createTestDocsTopic("id1",
				"www.somestorage.com/apiSpec.json",
				"www.somestorage.com/asyncApiSpec.json",
				"www.somestorage.com/content.json",
				docstopic.StatusNone)

			repositoryMock.On("Update", docsTopic).Return(nil)
		}

		{
			specFile := createTestInputFile(openApiSpecFileName, "", jsonApiSpec)
			eventsFile := createTestInputFile(eventsSpecFileName, "", eventsSpec)
			docsFile := createTestInputFile(documentationFileName, "", documentation)

			uploadClientMock.On("Upload", specFile).
				Return(createTestOutputFile(openApiSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", eventsFile).
				Return(createTestOutputFile(eventsSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", docsFile).
				Return(createTestOutputFile(documentationFileName, "www.somestorage.com"), nil)
		}

		// when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func createTestDocsTopic(id string, apiSpecUrl string, eventsSpecUrl string, documentationUrl string, status docstopic.StatusType) docstopic.Entry {

	return docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(DocTopicDescriptionFormat, id),
		Urls: map[string]string{
			docstopic.KeyOpenApiSpec:       apiSpecUrl,
			docstopic.KeyEventsSpec:        eventsSpecUrl,
			docstopic.KeyDocumentationSpec: documentationUrl,
		},
		Labels: map[string]string{DocsTopicLabelKey: DocsTopicLabelValue},
		Status: status,
	}
}

func createTestInputFile(name string, directory string, contents []byte) upload.File {
	return upload.File{
		Name:      name,
		Directory: directory,
		Contents:  contents,
	}
}

func createTestOutputFile(filename string, url string) upload.UploadedFile {
	return upload.UploadedFile{
		FileName:   filename,
		RemotePath: fmt.Sprintf("%s/%s", url, filename),
		Bucket:     "BucketName",
		Size:       100,
	}
}

func createTestServer(t *testing.T, body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(body)
		assert.NoError(t, err)

		w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
		w.WriteHeader(http.StatusOK)
	}))
}
