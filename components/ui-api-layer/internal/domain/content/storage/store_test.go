package storage_test

import (
	"bytes"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
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
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), errors.New("Random error"))

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
			Return(bytes.NewReader([]byte("{}")), nil)

		apiSpec, exists, err := service.ApiSpec("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func TestStore_AsyncApiSpec(t *testing.T) {
	t.Run("Not existing object", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "not-existing/asyncApiSpec.json").
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), errors.New("Random error"))

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
			Return(bytes.NewReader([]byte("{\"name\":\"test\",\"other\":\"yhm\"}")), nil)

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
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), nil)
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
			Return(bytes.NewReader([]byte{}), errors.New("Random error"))

		_, exists, err := service.Content("client-error")

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Docs with assets", func(t *testing.T) {
		client := storage.NewMockClient()
		service := storage.NewStore(client, "test", "https://test.ninja", "assets")

		client.On("Object", "test", "valid/content.json").
			Return(bytes.NewReader([]byte(fixContentWithLinksJSON())), nil)

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
			Return(bytes.NewReader([]byte("{}")), nil)

		apiSpec, exists, err := service.Content("valid")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
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
