package assetstore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload"
	uploadMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/upload/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const defaultAssetstoreRequestTimeout = 5

func TestAddingToAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")
	odataXMLApiSpec := []byte("<ODataServiceDocument xmlns:i=\"http://www.w3.org/2001/XMLSchema-instance\"" +
		"xmlns=\"http://schemas.datacontract.org/2004/07/Microsoft.OData.Core\"></ODataServiceDocument>")

	t.Run("Should put all specifications to asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		{
			urls := map[string]string{
				docstopic.KeyOpenApiSpec:       "www.somestorage.com/apiSpec.json",
				docstopic.KeyAsyncApiSpec:      "www.somestorage.com/asyncApiSpec.json",
				docstopic.KeyDocumentationSpec: "www.somestorage.com/content.json",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Upsert", docsTopic).Return(nil)
		}

		{
			uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
				Return(createUploadedFile(openApiSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", eventsSpecFileName, eventsSpec).
				Return(createUploadedFile(eventsSpecFileName, "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", documentationFileName, documentation).
				Return(createUploadedFile(documentationFileName, "www.somestorage.com"), nil)
		}

		// when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData XML specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Upsert", docsTopic).Return(nil)
		}

		uploadClientMock.On("Upload", odataXMLSpecFileName, odataXMLApiSpec).
			Return(createUploadedFile(odataXMLSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.ODataApiType, nil, odataXMLApiSpec, nil)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData JSON specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Upsert", docsTopic).Return(nil)
		}

		uploadClientMock.On("Upload", odataJSONSpecFileName, jsonApiSpec).
			Return(createUploadedFile(odataXMLSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.ODataApiType, nil, jsonApiSpec, nil)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(upload.UploadedFile{}, apperrors.Internal("some error"))

		// when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.Error(t, err)
		repositoryMock.AssertNotCalled(t, "Upsert")
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		repositoryMock.On("Upsert", mock.Anything).Return(apperrors.Internal("some error"))
		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(createUploadedFile(openApiSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.OpenApiType, nil, jsonApiSpec, nil)

		// then
		require.Error(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should not create DocsTopic if specs are not provided", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

		// when
		err := service.Put("id1", "", []byte(nil), []byte(nil), []byte(nil))

		// then
		assert.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
		repositoryMock.AssertNotCalled(t, "Upsert")
	})
}

func TestGettingFromAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should get specifications from asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		service := NewService(repositoryMock, nil, false, defaultAssetstoreRequestTimeout)

		apiTestServer := createTestServer(t, jsonApiSpec)
		defer apiTestServer.Close()

		eventTestServer := createTestServer(t, eventsSpec)
		defer eventTestServer.Close()

		documentationServer := createTestServer(t, documentation)
		defer documentationServer.Close()

		{
			urls := map[string]string{
				docstopic.KeyOpenApiSpec:       apiTestServer.URL,
				docstopic.KeyAsyncApiSpec:      eventTestServer.URL,
				docstopic.KeyDocumentationSpec: documentationServer.URL,
			}

			repositoryMock.On("Get", "id1").
				Return(createDocsTopic("id1", urls, docstopic.StatusReady), nil)
		}

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.NoError(t, err)
		assert.Equal(t, jsonApiSpec, api)
		assert.Equal(t, eventsSpec, events)
		assert.Equal(t, documentation, docs)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to read DocsTopic CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		service := NewService(repositoryMock, nil, false, defaultAssetstoreRequestTimeout)

		repositoryMock.On("Get", "id1").
			Return(docstopic.Entry{}, apperrors.Internal("some error"))

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.Error(t, err)
		assert.Equal(t, []byte(nil), api)
		assert.Equal(t, []byte(nil), events)
		assert.Equal(t, []byte(nil), docs)

		repositoryMock.AssertExpectations(t)
	})

	t.Run("Should return nil specs if Docs Topic CR doesn't exist", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		service := NewService(repositoryMock, nil, false, defaultAssetstoreRequestTimeout)

		repositoryMock.On("Get", "id1").
			Return(docstopic.Entry{}, apperrors.NotFound("object not found"))

		// then
		docs, api, events, err := service.Get("id1")

		// then
		require.NoError(t, err)
		assert.Equal(t, []byte(nil), api)
		assert.Equal(t, []byte(nil), events)
		assert.Equal(t, []byte(nil), docs)

		repositoryMock.AssertExpectations(t)
	})
}

func TestGettingFromAssetStoreIfStatusIsNotReady(t *testing.T) {
	statuses := []struct {
		description string
		status      docstopic.StatusType
	}{
		{"Should return nil specs if DocsTopic status is empty", docstopic.StatusNone},
		{"Should return nil specs if DocsTopic status is Failed", docstopic.StatusFailed},
		{"Should return nil specs if DocsTopic status is Pending", docstopic.StatusPending},
	}

	for _, testData := range statuses {
		t.Run(testData.description, func(t *testing.T) {
			// given
			repositoryMock := &mocks.DocsTopicRepository{}
			uploadClientMock := &uploadMocks.Client{}
			service := NewService(repositoryMock, uploadClientMock, false, defaultAssetstoreRequestTimeout)

			{
				repositoryMock.On("Get", "id1").Return(createDocsTopic("id1", nil, testData.status), nil)
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
}

func createDocsTopic(id string, urls map[string]string, status docstopic.StatusType) docstopic.Entry {
	return docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        urls,
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
		Status:      status,
	}
}

func createUploadedFile(filename string, url string) upload.UploadedFile {
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
