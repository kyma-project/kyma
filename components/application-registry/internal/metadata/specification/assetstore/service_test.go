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

	t.Run("Should add specifications to asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.Repository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		{
			docsTopic := createTestDocsTopic("id1",
				"www.somestorage.com/openapi.json",
				"www.somestorage.com/asyncapi.json",
				"www.somestorage.com/docs.json")

			repositoryMock.On("Update", docsTopic).Return(apperrors.NotFound("Not found."))
			repositoryMock.On("Create", docsTopic).Return(nil)
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
			&ContentEntry{"docs.json", docstopic.DocsTopicKeyDocumentationSpec, documentation},
			&ContentEntry{"openapi.json", docstopic.DocsTopicKeyOpenApiSpec, jsonApiSpec},
			&ContentEntry{"asyncapi.json", docstopic.DocsTopicKeyEventsSpec, eventsSpec},
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
	jsonApiSpec := []byte("{\"productsEndpoint\": \"Endpoint /products returns products.\"}}")
	documentation := []byte("{\"description\": \"Some docs blah blah blah\"}}")
	eventsSpec := []byte("{\"orderCreated\": \"Published when order is placed.\"}}")

	t.Run("Should get specifications from asset store", func(t *testing.T) {
		// given
		repositoryMock := &mocks.DocsTopicRepository{}
		uploadClientMock := &uploadMocks.Client{}
		service := NewService(repositoryMock, uploadClientMock)

		apiTestServer := createTestServer(jsonApiSpec)
		defer apiTestServer.Close()

		eventTestServer := createTestServer(eventsSpec)
		defer eventTestServer.Close()

		documentationServer := createTestServer(documentation)
		defer documentationServer.Close()

		{
			repositoryMock.On("Get", "id1").Return(docstopic.Entry{
				Id:          "id1",
				DisplayName: "Some display name",
				Description: "Some description",
				Urls: map[string]string{
					docstopic.DocsTopicKeyOpenApiSpec:       apiTestServer.URL,
					docstopic.DocsTopicKeyEventsSpec:        eventTestServer.URL,
					docstopic.DocsTopicKeyDocumentationSpec: documentationServer.URL,
				},
			}, nil)
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

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

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
				"www.somestorage.com/openapi.json",
				"www.somestorage.com/asyncapi.json",
				"www.somestorage.com/docs.json")

			repositoryMock.On("Update", docsTopic).Return(nil)
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
			&ContentEntry{"docs.json", docstopic.DocsTopicKeyDocumentationSpec, documentation},
			&ContentEntry{"openapi.json", docstopic.DocsTopicKeyOpenApiSpec, jsonApiSpec},
			&ContentEntry{"asyncapi.json", docstopic.DocsTopicKeyEventsSpec, eventsSpec},
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

func TestRemovingFromAssetStore(t *testing.T) {
	t.Run("Should remove specifications from asset store", func(t *testing.T) {

	})

	t.Run("Should fail when failed to upload file", func(t *testing.T) {

	})

	t.Run("Should fail when failed to create DocsTopic CR", func(t *testing.T) {

	})
}

func createTestDocsTopic(id string, apiSpecUrl string, eventsSpecUrl string, documentationUrl string) docstopic.Entry {

	return docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(DocTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(DocTopicDescriptionFormat, id),
		Urls: map[string]string{
			docstopic.DocsTopicKeyOpenApiSpec:       apiSpecUrl,
			docstopic.DocsTopicKeyEventsSpec:        eventsSpecUrl,
			docstopic.DocsTopicKeyDocumentationSpec: documentationUrl,
		},
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

func createTestServer(body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
		w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
		w.WriteHeader(http.StatusOK)
	}))
}
