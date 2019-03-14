package storage_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStore_ApiSpec(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(true)

		_, exists, err := service.ApiSpec("not-existing")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "invalid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("?<"))), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(false)

		_, exists, err := service.ApiSpec("invalid")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Client error", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "client-error/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("Random error"))

		_, exists, err := service.ApiSpec("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Valid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		expected := &storage.ApiSpec{
			Raw: map[string]interface{}{},
		}

		client.On("Object", "test", "valid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("{}"))), nil)

		apiSpec, exists, err := service.ApiSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func TestStore_OpenApiSpec(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(true)

		_, exists, err := service.OpenApiSpec("not-existing")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "invalid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("?<"))), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(false)

		_, exists, err := service.OpenApiSpec("invalid")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Client error", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "client-error/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("Random error"))

		_, exists, err := service.OpenApiSpec("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Valid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		expected := &storage.OpenApiSpec{
			Raw: map[string]interface{}{},
		}

		client.On("Object", "test", "valid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("{}"))), nil)

		openApiSpec, exists, err := service.OpenApiSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, openApiSpec)
	})
}

func TestStore_ODataApiSpec(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(true)

		_, exists, err := service.ODataSpec("not-existing")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "invalid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("?<"))), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(false)

		_, exists, err := service.ODataSpec("invalid")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Client error", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "client-error/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("Random error"))

		_, exists, err := service.ODataSpec("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid json object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		expected := &storage.ODataSpec{
			Raw: "",
		}

		client.On("Object", "test", "invalid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("{}"))), nil)

		odataSpec, exists, err := service.ODataSpec("invalid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})

	t.Run("Valid json object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		jsonExample := fixODataJSONV4()
		expected := &storage.ODataSpec{
			Raw: jsonExample,
		}

		client.On("Object", "test", "valid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte(jsonExample))), nil)

		odataSpec, exists, err := service.ODataSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})

	t.Run("Valid xml object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		xmlExample := fixODataXML()
		expected := &storage.ODataSpec{
			Raw: xmlExample,
		}

		client.On("Object", "test", "valid/apiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte(xmlExample))), nil)

		odataSpec, exists, err := service.ODataSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})
}

func TestStore_AsyncApiSpec(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/asyncApiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(true)

		_, exists, err := service.AsyncApiSpec("not-existing")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "invalid/asyncApiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("<>"))), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(false)

		_, exists, err := service.AsyncApiSpec("invalid")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Client error", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "client-error/asyncApiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("Random error"))

		_, exists, err := service.AsyncApiSpec("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Valid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		expected := &storage.AsyncApiSpec{
			Data: storage.AsyncApiSpecData{},
			Raw: map[string]interface{}{
				"name":  "test",
				"other": "yhm",
			},
		}

		client.On("Object", "test", "valid/asyncApiSpec.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("{\"name\":\"test\",\"other\":\"yhm\"}"))), nil)

		apiSpec, exists, err := service.AsyncApiSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func TestStore_Content(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/content.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(true)

		_, exists, err := service.Content("not-existing")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Invalid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "invalid/content.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("<>"))), nil)
		client.On("IsNotExistsError", mock.Anything).
			Return(false)

		_, exists, err := service.Content("invalid")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Client error", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "client-error/content.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte{})), errors.New("Random error"))

		_, exists, err := service.Content("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Docs with assets", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "valid/content.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte(fixContentWithLinksJSON()))), nil)

		apiSpec, exists, err := service.Content("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, fixContentWithLinks(), apiSpec)
	})

	t.Run("Valid object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		expected := &storage.Content{
			Raw: map[string]interface{}{},
		}

		client.On("Object", "test", "valid/content.json").
			Return(ioutil.NopCloser(bytes.NewReader([]byte("{}"))), nil)

		apiSpec, exists, err := service.Content("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func fixODataJSONV4() string {
	return `
{
  "$Version": "4.01",
  "$EntityContainer": "ODataDemo.DemoService"
}
`
}

func fixODataXML() string {
	return `
<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx Version="4.0" xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx">
  <edmx:Reference Uri="http://tinyurl.com/Org-OData-Core">
    <edmx:Include Namespace="Org.OData.Core.V1" Alias="Core" />
  </edmx:Reference>
</edmx:Edmx>
`
}

func fixContentWithLinksJSON() string {
	return `{
 "displayName": "Example Docs",
 "id": "example-docs",
 "type": "Service Class",
 "description": "Foo bar baz",
 "docs": [
   {
     "title": "With placeholder",
     "type": "doctype1",
     "source": "<img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />"
   },
   {
     "title": "With dot",
     "type": "doctype2",
     "source": "<img src=\"./assets/image.jpg\" />"
   },
   {
     "title": "Without dot",
     "type": "doctype3",
     "source": "<img src=\"assets/image.jpg\" />"
   },
   {
     "title": "Mixed",
     "type": "doctype3",
     "source": "<img src=\"assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />"
   },
   {
     "title": "Mixed multiple",
     "type": "doctype3",
     "source": "<img src=\"assets/image.jpg\" /><img src=\"assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />"
   }
 ]
}`
}

func fixContentWithLinks() *storage.Content {
	return &storage.Content{
		Raw: map[string]interface{}{
			"displayName": "Example Docs",
			"id":          "example-docs",
			"type":        "Service Class",
			"description": "Foo bar baz",
			"docs": []interface{}{
				map[string]interface{}{
					"title":  "With placeholder",
					"type":   "doctype1",
					"source": "<img src=\"https://test.ninja/service-class/example-docs/assets/image.jpg\" />",
				},
				map[string]interface{}{
					"title":  "With dot",
					"type":   "doctype2",
					"source": "<img src=\"https://test.ninja/test/valid/assets/image.jpg\" />",
				},
				map[string]interface{}{
					"title":  "Without dot",
					"type":   "doctype3",
					"source": "<img src=\"https://test.ninja/test/valid/assets/image.jpg\" />",
				},
				map[string]interface{}{
					"title":  "Mixed",
					"type":   "doctype3",
					"source": "<img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/service-class/example-docs/assets/image.jpg\" />",
				},
				map[string]interface{}{
					"title":  "Mixed multiple",
					"type":   "doctype3",
					"source": "<img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/test/valid/assets/image.jpg\" /><img src=\"https://test.ninja/service-class/example-docs/assets/image.jpg\" /><img src=\"https://test.ninja/service-class/example-docs/assets/image.jpg\" />",
				},
			},
		},
		Data: storage.ContentData{
			Description: "Foo bar baz",
			DisplayName: "Example Docs",
			Type:        "Service Class",
			Docs: []storage.Document{
				{
					Order:    "",
					Source:   "<img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />",
					Title:    "With placeholder",
					Type:     "doctype1",
					Internal: false,
				},
				{
					Order:    "",
					Source:   "<img src=\"./assets/image.jpg\" />",
					Title:    "With dot",
					Type:     "doctype2",
					Internal: false,
				},
				{
					Order:    "",
					Source:   "<img src=\"assets/image.jpg\" />",
					Title:    "Without dot",
					Type:     "doctype3",
					Internal: false,
				},
				{
					Order:    "",
					Source:   "<img src=\"assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />",
					Title:    "Mixed",
					Type:     "doctype3",
					Internal: false,
				},
				{
					Order:    "",
					Source:   "<img src=\"assets/image.jpg\" /><img src=\"assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"./assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" /><img src=\"{PLACEHOLDER_APP_RESOURCES_BASE_URI}/service-class/example-docs/assets/image.jpg\" />",
					Title:    "Mixed multiple",
					Type:     "doctype3",
					Internal: false,
				},
			},
			ID: "example-docs",
		},
	}
}
