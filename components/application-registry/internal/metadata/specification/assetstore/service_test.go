package assetstore

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/docstopic"
	docsTopicMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/docstopic/mocks"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/upload"
	uploadMocks "github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/upload/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddingToAssetStore(t *testing.T) {
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should add specifications to asset store", func(t *testing.T) {
		// given
		repositoryMock := &docsTopicMocks.Repository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			matcher := getDocumentationTopicMatcher(
				"www.somestorage.com/openapi.json",
				"www.somestorage.com/asyncapi.json",
				"www.somestorage.com/docs.json")
			repositoryMock.On("Update", mock.MatchedBy(matcher)).Return(apperrors.NotFound("Not found."))
			repositoryMock.On("Create", mock.MatchedBy(matcher)).Return(nil)
		}

		{
			specFile := createTestInputFile("openapi.json", "id1", jsonApiSpec)
			eventsFile := createTestInputFile("asyncapi.json", "id1", eventsSpec)
			docsFile := createTestInputFile("docs.json", "id1", documentation)

			uploadClientMock.On("Upload", specFile).
				Return(createTestOutputFile("openapi.json", "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", eventsFile).
				Return(createTestOutputFile("asyncapi.json", "www.somestorage.com"), nil)

			uploadClientMock.On("Upload", docsFile).
				Return(createTestOutputFile("docs.json", "www.somestorage.com"), nil)
		}

		// when
		err := service.Put("id1",
			ContentEntry{"docs.json", documentation},
			ContentEntry{"openapi.json", jsonApiSpec},
			ContentEntry{"asyncapi.json", eventsSpec},
		)
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

func TestGettingFromAssetStore(t *testing.T) {
	//jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	//documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	//eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should get specifications from asset store", func(t *testing.T) {
		//	// given
		//	repositoryMock := &docsTopicMocks.Repository{}
		//	uploadClientMock := &uploadMocks.Client{}
		//	service := NewService(repositoryMock, uploadClientMock)
		//
		//	{
		//		repositoryMock.On("Get", "id1").Return(docstopic.DocumentationTopic{
		//			Id:               "id1",
		//			DisplayName:      "Some display name",
		//			Description:      "Some description",
		//			ApiSpecUrl:       "some url",
		//			EventsSpecUrl:    "some url",
		//			DocumentationUrl: "some url",
		//		}, nil)
		//	}
		//
		//	// then
		//	api, events, docs, err := service.Get("id1")
		//
		//	// then
		//	require.NoError(t, err)
		//	assert.Equal(t, jsonApiSpec, api)
		//	assert.Equal(t, eventsSpec, events)
		//	assert.Equal(t, documentation, docs)
		//
		//	repositoryMock.AssertExpectations(t)
		//	uploadClientMock.AssertExpectations(t)
	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func TestUpdatingInAssetStore(t *testing.T) {
	t.Run("Should update specifications in asset store", func(t *testing.T) {

	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func TestRemovingFromAssetStore(t *testing.T) {
	t.Run("Should remove specifications from asset store", func(t *testing.T) {

	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func getDocumentationTopicMatcher(apiSpecUrl string, eventsSpecUrl string, documentationUrl string) func(docstopic.DocumentationTopic) bool {
	return func(d docstopic.DocumentationTopic) bool {

		descriptionNonEmpty := len(d.DisplayName) > 0 && len(d.DisplayName) > 0
		specsMatch := d.ApiSpecUrl == apiSpecUrl &&
			d.EventsSpecUrl == eventsSpecUrl &&
			d.DocumentationUrl == documentationUrl

		return descriptionNonEmpty && specsMatch
	}
}

func createTestInputFile(name string, directory string, contents []byte) upload.InputFile {
	return upload.InputFile{
		Name:      name,
		Directory: directory,
		Contents:  contents,
	}
}

func createTestOutputFile(filename string, url string) upload.OutputFile {
	return upload.OutputFile{
		FileName:   filename,
		RemotePath: fmt.Sprintf("%s/%s", url, filename),
		Bucket:     "BucketName",
		Size:       100,
	}
}
