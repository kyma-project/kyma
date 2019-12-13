package rafter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/clusterassetgroup"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/upload"
	uploadMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/upload/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const defaultRafterRequestTimeout = 5

func TestAddingToRafter(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")
	odataXMLApiSpec := []byte("<ODataServiceDocument xmlns:i=\"http://www.w3.org/2001/XMLSchema-instance\"" +
		"xmlns=\"http://schemas.datacontract.org/2004/07/Microsoft.OData.Core\"></ODataServiceDocument>")

	t.Run("Should put all specifications to rafter", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		{
			urls := map[string]string{
				clusterassetgroup.KeyOpenApiSpec:       "www.somestorage.com/apiSpec.json",
				clusterassetgroup.KeyAsyncApiSpec:      "www.somestorage.com/asyncApiSpec.json",
				clusterassetgroup.KeyDocumentationSpec: "www.somestorage.com/content.json",
			}
			clusterAssetGroup := createClusterAssetGroup("id1", urls, clusterassetgroup.StatusNone)

			repositoryMock.On("Upsert", clusterAssetGroup).Return(nil)
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
		err := service.Put("id1", clusterassetgroup.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData XML specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		{
			urls := map[string]string{
				clusterassetgroup.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			clusterAssetGroup := createClusterAssetGroup("id1", urls, clusterassetgroup.StatusNone)

			repositoryMock.On("Upsert", clusterAssetGroup).Return(nil)
		}

		uploadClientMock.On("Upload", odataXMLSpecFileName, odataXMLApiSpec).
			Return(createUploadedFile(odataXMLSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", clusterassetgroup.ODataApiType, nil, odataXMLApiSpec, nil)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should detect OData JSON specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		{
			urls := map[string]string{
				clusterassetgroup.KeyODataSpec: "www.somestorage.com/odata.xml",
			}
			clusterAssetGroup := createClusterAssetGroup("id1", urls, clusterassetgroup.StatusNone)

			repositoryMock.On("Upsert", clusterAssetGroup).Return(nil)
		}

		uploadClientMock.On("Upload", odataJSONSpecFileName, jsonApiSpec).
			Return(createUploadedFile(odataXMLSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", clusterassetgroup.ODataApiType, nil, jsonApiSpec, nil)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(upload.UploadedFile{}, apperrors.Internal("some error"))

		// when
		err := service.Put("id1", clusterassetgroup.OpenApiType, documentation, jsonApiSpec, eventsSpec)

		// then
		require.Error(t, err)
		repositoryMock.AssertNotCalled(t, "Upsert")
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to create ClusterAssetGroup CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		repositoryMock.On("Upsert", mock.Anything).Return(apperrors.Internal("some error"))
		uploadClientMock.On("Upload", openApiSpecFileName, jsonApiSpec).
			Return(createUploadedFile(openApiSpecFileName, "www.somestorage.com"), nil)

		// when
		err := service.Put("id1", clusterassetgroup.OpenApiType, nil, jsonApiSpec, nil)

		// then
		require.Error(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should not create ClusterAssetGroup if specs are not provided", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

		// when
		err := service.Put("id1", "", []byte(nil), []byte(nil), []byte(nil))

		// then
		assert.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
		repositoryMock.AssertNotCalled(t, "Upsert")
	})
}

func TestGettingFromRafter(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should get specifications from rafter", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		service := NewService(repositoryMock, nil, false, defaultRafterRequestTimeout)

		apiTestServer := createTestServer(t, jsonApiSpec)
		defer apiTestServer.Close()

		eventTestServer := createTestServer(t, eventsSpec)
		defer eventTestServer.Close()

		documentationServer := createTestServer(t, documentation)
		defer documentationServer.Close()

		{
			urls := map[string]string{
				clusterassetgroup.KeyOpenApiSpec:       apiTestServer.URL,
				clusterassetgroup.KeyAsyncApiSpec:      eventTestServer.URL,
				clusterassetgroup.KeyDocumentationSpec: documentationServer.URL,
			}

			repositoryMock.On("Get", "id1").
				Return(createClusterAssetGroup("id1", urls, clusterassetgroup.StatusReady), nil)
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

	t.Run("Should fail when failed to read ClusterAssetGroup CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		service := NewService(repositoryMock, nil, false, defaultRafterRequestTimeout)

		repositoryMock.On("Get", "id1").
			Return(clusterassetgroup.Entry{}, apperrors.Internal("some error"))

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
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		service := NewService(repositoryMock, nil, false, defaultRafterRequestTimeout)

		repositoryMock.On("Get", "id1").
			Return(clusterassetgroup.Entry{}, apperrors.NotFound("object not found"))

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

func TestGettingFromRafterIfStatusIsNotReady(t *testing.T) {
	statuses := []struct {
		description string
		status      clusterassetgroup.StatusType
	}{
		{"Should return nil specs if ClusterAssetGroup status is empty", clusterassetgroup.StatusNone},
		{"Should return nil specs if ClusterAssetGroup status is Failed", clusterassetgroup.StatusFailed},
		{"Should return nil specs if ClusterAssetGroup status is Pending", clusterassetgroup.StatusPending},
	}

	for _, testData := range statuses {
		t.Run(testData.description, func(t *testing.T) {
			// given
			repositoryMock := &mocks.ClusterAssetGroupRepository{}
			uploadClientMock := &uploadMocks.Client{}
			service := NewService(repositoryMock, uploadClientMock, false, defaultRafterRequestTimeout)

			{
				repositoryMock.On("Get", "id1").Return(createClusterAssetGroup("id1", nil, testData.status), nil)
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

func createClusterAssetGroup(id string, urls map[string]string, status clusterassetgroup.StatusType) clusterassetgroup.Entry {
	return clusterassetgroup.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(clusterAssetGroupNameFormat, id),
		Description: fmt.Sprintf(clusterAssetGroupDescriptionFormat, id),
		Urls:        urls,
		Labels:      map[string]string{clusterAssetGroupLabelKey: clusterAssetGroupLabelValue},
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
