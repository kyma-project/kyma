package rafter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/clusterassetgroup"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/mocks"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/upload"
	uploadMocks "kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/rafter/upload/mocks"
)

func TestAddingToRafter(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")
	odataXMLApiSpec := []byte("<ODataServiceDocument xmlns:i=\"http://www.w3.org/2001/XMLSchema-instance\"" +
		"xmlns=\"http://schemas.datacontract.org/2004/07/Microsoft.OData.Core\"></ODataServiceDocument>")

	specFormatJSON := clusterassetgroup.SpecFormatJSON
	specFormatXML := clusterassetgroup.SpecFormatXML

	inputAPIAsset := clusterassetgroup.Asset{
		Name:    "assetId",
		Type:    clusterassetgroup.OpenApiType,
		Format:  specFormatJSON,
		Content: jsonApiSpec,
	}

	expectedAPIAsset := clusterassetgroup.Asset{
		Name:     "assetId",
		Type:     clusterassetgroup.OpenApiType,
		Format:   specFormatJSON,
		Content:  jsonApiSpec,
		SpecHash: "d6398860c75774aef6ee4fc97f8b8d44",
		Url:      "www.somestorage.com/apiSpec.json",
	}

	inputEventsAPIAsset := clusterassetgroup.Asset{
		Name:    "assetId",
		Type:    clusterassetgroup.AsyncApi,
		Format:  specFormatJSON,
		Content: eventsSpec,
	}

	expectedAEventsPIAsset := clusterassetgroup.Asset{
		Name:     "assetId",
		Type:     clusterassetgroup.AsyncApi,
		Format:   specFormatJSON,
		Content:  eventsSpec,
		SpecHash: "a83a71e0f48551fab2f079805b39ebd1",
		Url:      "www.somestorage.com/asyncApiSpec",
	}

	t.Run("Should put api and event spec to rafter", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		expectedAPIEntry := clusterassetgroup.Entry{
			Id:          "id1",
			DisplayName: "Documentation topic for service class id=id1",
			Description: "Documentation topic for service class id=id1",
			Assets:      []clusterassetgroup.Asset{expectedAPIAsset, expectedAEventsPIAsset},
			Labels: map[string]string{
				"rafter.kyma-project.io/view-context": "service-catalog",
			},
		}

		repositoryMock.On("Get", "id1").Return(clusterassetgroup.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", expectedAPIEntry).Return(nil)

		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(openApiSpecFileName, specFormatJSON), "www.somestorage.com"), nil)

		uploadClientMock.On("Upload", specFileName(eventsSpecFileName, specFormatJSON), eventsSpec).
			Return(createUploadedFile(eventsSpecFileName, "www.somestorage.com"), nil)

		assets := []clusterassetgroup.Asset{inputAPIAsset, inputEventsAPIAsset}
		// when
		err := service.Put("id1", assets)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should handle OData XML specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(clusterassetgroup.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(nil)

		uploadClientMock.On("Upload", specFileName(odataSpecFileName, specFormatXML), odataXMLApiSpec).
			Return(createUploadedFile(specFileName(odataSpecFileName, specFormatXML), "www.somestorage.com"), nil)

		assets := []clusterassetgroup.Asset{
			{
				Name:    "assetId",
				Type:    clusterassetgroup.ODataApiType,
				Format:  specFormatXML,
				Content: odataXMLApiSpec,
			},
		}

		// when
		err := service.Put("id1", assets)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should handle OData JSON specification", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(clusterassetgroup.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(nil)

		uploadClientMock.On("Upload", specFileName(odataSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(odataSpecFileName, specFormatXML), "www.somestorage.com"), nil)

		assets := []clusterassetgroup.Asset{
			{
				Name:    "assetId",
				Type:    clusterassetgroup.ODataApiType,
				Format:  specFormatJSON,
				Content: jsonApiSpec,
			},
		}

		// when
		err := service.Put("id1", assets)

		// then
		require.NoError(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(clusterassetgroup.Entry{}, apperrors.NotFound("Not found"))

		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(upload.UploadedFile{}, apperrors.Internal("some error"))

		assets := []clusterassetgroup.Asset{inputAPIAsset}

		// when
		err := service.Put("id1", assets)

		// then
		require.Error(t, err)
		repositoryMock.AssertNotCalled(t, "Upsert")
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to create ClusterAssetGroup CR", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		repositoryMock.On("Get", "id1").Return(clusterassetgroup.Entry{}, apperrors.NotFound("Not found"))
		repositoryMock.On("Create", mock.Anything).Return(apperrors.Internal("some error"))
		uploadClientMock.On("Upload", specFileName(openApiSpecFileName, specFormatJSON), jsonApiSpec).
			Return(createUploadedFile(specFileName(openApiSpecFileName, specFormatJSON), "www.somestorage.com"), nil)

		assets := []clusterassetgroup.Asset{inputAPIAsset}

		// when
		err := service.Put("id1", assets)

		// then
		require.Error(t, err)
		repositoryMock.AssertExpectations(t)
		uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should not update ClusterAssetGroup when provided spec is identical with stored one", func(t *testing.T) {
		//given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		storedEntry := clusterassetgroup.Entry{
			Id:          "id1",
			DisplayName: fmt.Sprintf(clusterAssetGroupDisplayNameFormat, "id1"),
			Description: fmt.Sprintf(clusterAssetGroupDescriptionFormat, "id1"),
			Labels:      map[string]string{clusterAssetGroupLabelKey: clusterAssetGroupLabelValue},
			Status:      clusterassetgroup.StatusNone,
			Assets:      []clusterassetgroup.Asset{expectedAPIAsset, expectedAEventsPIAsset},
		}

		repositoryMock.On("Get", "id1").Return(storedEntry, nil)

		assets := []clusterassetgroup.Asset{inputAPIAsset, inputEventsAPIAsset}

		//when
		err := service.Put("id1", assets)

		// then
		require.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
	})

	t.Run("Should not create ClusterAssetGroup if specs is not provided", func(t *testing.T) {
		// given
		repositoryMock := &mocks.ClusterAssetGroupRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		// when
		err := service.Put("id1", []clusterassetgroup.Asset{})

		// then
		assert.NoError(t, err)
		uploadClientMock.AssertNotCalled(t, "Upload")
		repositoryMock.AssertNotCalled(t, "Get")
		repositoryMock.AssertNotCalled(t, "Create")
		repositoryMock.AssertNotCalled(t, "Update")
	})
}

func createUploadedFile(filename string, url string) upload.UploadedFile {
	return upload.UploadedFile{
		FileName:   filename,
		RemotePath: fmt.Sprintf("%s/%s", url, filename),
		Bucket:     "BucketName",
		Size:       100,
	}
}
