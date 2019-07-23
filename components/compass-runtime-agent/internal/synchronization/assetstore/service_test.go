package assetstore

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/httpconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/upload"
	uploadMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/assetstore/upload/mocks"
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
