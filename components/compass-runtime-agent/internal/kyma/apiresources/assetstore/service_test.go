package assetstore

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload"
	uploadMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
		service := NewService(repositoryMock, uploadClientMock)

		{
			urls := map[string]string{
				docstopic.KeyOpenApiSpec:       "www.somestorage.com/apiSpec.json",
				docstopic.KeyAsyncApiSpec:      "www.somestorage.com/asyncApiSpec.json",
				docstopic.KeyDocumentationSpec: "www.somestorage.com/content.json",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
			repositoryMock.On("Create", mock.Anything).Return(nil)
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
		service := NewService(repositoryMock, uploadClientMock)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
			repositoryMock.On("Create", mock.Anything).Return(nil)
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
		service := NewService(repositoryMock, uploadClientMock)

		{
			urls := map[string]string{
				docstopic.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			docsTopic := createDocsTopic("id1", urls, docstopic.StatusNone)

			repositoryMock.On("Get", docsTopic.Id).Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
			repositoryMock.On("Create", mock.Anything).Return(nil)
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
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(docstopic.Entry{}, apperrors.NotFound("Not found"))

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
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(docstopic.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(apperrors.Internal("some error"))
		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(createUploadedFile(openApiSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", docstopic.OpenApiType, nil, jsonApiSpec, nil)

		// then
		require.Error(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should upload only those files that are not identical with stored ones", func(t *testing.T) {
		//given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		urls := map[string]string{
			docstopic.KeyOpenApiSpec:       "www.somestorage.com/apiSpec.json",
			docstopic.KeyAsyncApiSpec:      "www.somestorage.com/asyncApiSpec.json",
			docstopic.KeyDocumentationSpec: "www.somestorage.com/content.json",
		}

		storedJsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /stored apiSpec.\"}}")

		storedEntry := createDocsTopicWithHashes("id1", urls, docstopic.StatusNone, storedJsonApiSpec, eventsSpec, documentation)

		repositoryMock.On("Get", "id1").Return(storedEntry, nil)
		repositoryMock.On("Update", mock.Anything).Return(nil)

		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(createUploadedFile(openApiSpecFileName, "www.somestorage.com"), nil)

		//when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		uploadClientMock.AssertNumberOfCalls(t, "Upload", 1)
	})

	t.Run("Should not update DocsTopic when provided specs are identical with stored ones", func(t *testing.T) {
		//given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		urls := map[string]string{
			docstopic.KeyOpenApiSpec:       "www.somestorage.com/apiSpec.json",
			docstopic.KeyAsyncApiSpec:      "www.somestorage.com/asyncApiSpec.json",
			docstopic.KeyDocumentationSpec: "www.somestorage.com/content.json",
		}
		storedEntry := createDocsTopicWithHashes("id1", urls, docstopic.StatusNone, jsonApiSpec, eventsSpec, documentation)

		repositoryMock.On("Get", "id1").Return(storedEntry, nil)
		repositoryMock.On("Update", mock.Anything).Return(nil)

		//when
		err := service.Put("id1", docstopic.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
	})

	t.Run("Should not create DocsTopic if specs are not provided", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		// when
		err := service.Put("id1", "", []byte(nil), []byte(nil), []byte(nil))

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

func createDocsTopicWithHashes(id string, urls map[string]string, status docstopic.StatusType, apiSpec, eventsSpec, documentation []byte) docstopic.Entry {
	return docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        urls,
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
		Status:      status,
		Hashes:      calculateHashes(documentation, apiSpec, eventsSpec),
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
