package rafter_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/rafter"
)

func TestSpecificationService_AsyncAPI(t *testing.T) {
	endpoint := "http://example.com"

	t.Run("Failed", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 500, "")
		data, err := service.AsyncAPI("foo/bucket/asset", "bar.yaml")

		require.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("Nil", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 200, "")
		data, err := service.AsyncAPI("foo/bucket/asset", "bar.yaml")

		require.NoError(t, err)
		assert.Nil(t, data)
	})
}

func TestSpecificationService_ReadData(t *testing.T) {
	endpoint := "http://example.com"
	body := "Foo-Bar"

	t.Run("Success", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 200, body)
		data, err := service.ReadData("foo/bucket/asset", "bar.yaml")

		require.NoError(t, err)
		assert.Equal(t, []byte("Foo-Bar"), data)
	})

	t.Run("Failed", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 500, "")
		data, err := service.ReadData("foo/bucket/asset", "bar.yaml")

		require.Error(t, err)
		assert.Nil(t, data)
	})

	t.Run("Nil (empty baseURL and name parameters)", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 200, "")
		data, err := service.ReadData("", "")

		require.NoError(t, err)
		assert.Nil(t, data)
	})

	t.Run("Nil (empty response)", func(t *testing.T) {
		service := fixSpecificationService(endpoint, 200, "")
		data, err := service.ReadData("foo/bucket/asset", "bar.yaml")

		require.NoError(t, err)
		assert.Nil(t, data)
	})
}

func TestSpecificationService_PreparePath(t *testing.T) {
	endpoint := "http://example.com"
	service := fixSpecificationService(endpoint, 200, "")

	t.Run("Success", func(t *testing.T) {
		path := service.PreparePath("foo/bucket/asset", "bar.yaml")
		expected := fmt.Sprintf("%s/%s", endpoint, "bucket/asset/bar.yaml")
		assert.Equal(t, expected, path)
	})

	t.Run("A small number of elements after split baseURL parameter", func(t *testing.T) {
		path := service.PreparePath("foo/bar", "foo")
		assert.Empty(t, path)
	})

	t.Run("Empty baseURL parameter", func(t *testing.T) {
		path := service.PreparePath("", "foo")
		assert.Empty(t, path)
	})

	t.Run("Empty name parameter", func(t *testing.T) {
		path := service.PreparePath("foo", "")
		assert.Empty(t, path)
	})

	t.Run("Empty baseURL and name parameters", func(t *testing.T) {
		path := service.PreparePath("", "")
		assert.Empty(t, path)
	})
}

func TestSpecificationService_Fetch(t *testing.T) {
	url := "http://example.com/bucket/asset/bar.yaml"
	body := "Foo-Bar"

	t.Run("Success", func(t *testing.T) {
		service := fixSpecificationService(url, 200, body)
		data, err := service.Fetch(url)

		require.NoError(t, err)
		assert.Equal(t, []byte("Foo-Bar"), data)
	})

	t.Run("Failed", func(t *testing.T) {
		service := fixSpecificationService(url, 500, "")
		data, err := service.Fetch(url)

		require.Error(t, err)
		assert.Nil(t, data)
	})
}

func fixSpecificationService(endpoint string, statusCode int, body string) *rafter.SpecificationService {
	return rafter.NewSpecificationService(rafter.Config{}, endpoint, newFakeHttpClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: statusCode,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(body)),
			// Must be set to non-nil value or it panics
			Header: make(http.Header),
		}
	}))
}

type roundTrip func(req *http.Request) *http.Response

func (f roundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func newFakeHttpClient(fn roundTrip) *http.Client {
	return &http.Client{
		Transport: roundTrip(fn),
	}
}
