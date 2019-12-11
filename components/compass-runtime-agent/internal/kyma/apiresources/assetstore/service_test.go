package assetstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload"
	uploadMocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload/mocks"
)

func TestAddingToAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")
	odataXMLApiSpec := []byte("<ODataServiceDocument xmlns:i=\"http://www.w3.org/2001/XMLSchema-instance\"" +
		"xmlns=\"http://schemas.datacontract.org/2004/07/Microsoft.OData.Core\"></ODataServiceDocument>")

	specFormatJSON := docstopic.SpecFormatJSON
	specFormatXML := docstopic.SpecFormatXML

	t.Run("Should put api spec to asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		urls := map[string]string{
			docstopic.KeyOpenApiSpec: "www.somestorage.com/apiSpec.json",
		}
		docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

		repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(nil)

		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(openApiSpecFileName, specFormatJSON), "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should put event api spec to asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		urls := map[string]string{
			docstopic.KeyAsyncApiSpec: "www.somestorage.com/asyncApiSpec.json",
		}
		docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

		repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(nil)

		uploadClientMock.On("Upload", specFileName(eventsSpecFileName, specFormatJSON), eventsSpec).
			Return(createUploadedFile(eventsSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.AsyncApi, eventsSpec, specFormatJSON, docstopic.EventApiSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData XML specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
			repositoryMock.On("Create", mock.Anything).Return(nil)
		}

		uploadClientMock.On("Upload", specFileName(odataSpecFileName, specFormatXML), odataXMLApiSpec).
			Return(createUploadedFile(specFileName(odataSpecFileName, specFormatXML), "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.ODataApiType, odataXMLApiSpec, specFormatXML, docstopic.ApiSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData JSON specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
			repositoryMock.On("Create", mock.Anything).Return(nil)
		}

		uploadClientMock.On("Upload", specFileName(odataSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(odataSpecFileName, specFormatXML), "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.ODataApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(docstopic.Entry{}, apperrors.NotFound("Not found"))

		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(upload.UploadedFile{}, apperrors.Internal("some error"))

		// when
		err := service.Put("id1", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec)

		// then
		require.Error(t, err)
		repositoryMock.AssertNotCalled(t, "Upsert")
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(apperrors.Internal("some error"))
		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(openApiSpecFileName, specFormatJSON), "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec)

		// then
		require.Error(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should not update DocsTopic when provided spec is identical with stored one", func(t *testing.T) {
		//given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		urls := map[string]string{
			docstopic.KeyOpenApiSpec: "www.somestorage.com/apiSpec.json",
		}
		storedEntry := createDocsTopicWithHashes("id1", urls, docstopic.StatusNone, jsonApiSpec)

		repositoryMock.On("Get", "id1").Return(storedEntry, nil)
		repositoryMock.On("Update", mock.Anything).Return(nil)

		//when
		err := service.Put("id1", docstopic.OpenApiType, jsonApiSpec, specFormatJSON, docstopic.ApiSpec)

		// then
		require.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
	})

	t.Run("Should not create DocsTopic if specs is not provided", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		// when
		err := service.Put("id1", "", []byte(nil), specFormatJSON, docstopic.ApiSpec)

		// then
		assert.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
		repositoryMock.AssertNotCalled(t, "Get")
		repositoryMock.AssertNotCalled(t, "Create")
		repositoryMock.AssertNotCalled(t, "Update")
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

func createDocsTopicWithHashes(id string, urls map[string]string, status docstopic.StatusType, apiSpec []byte) docstopic.Entry {
	return docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        urls,
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
		Status:      status,
		SpecHash:    calculateHash(apiSpec),
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
