package externalapi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiSpecHandler_HandleRequest(t *testing.T) {
	t.Run("should serve specified file", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/api.yaml", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		handler := NewStaticFileHandler("testdata/apispec.yaml")

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, rr.HeaderMap.Get("Content-Type"), "text/plain; charset=utf-8")

		b, err := ioutil.ReadFile("testdata/apispec.yaml")
		require.NoError(t, err)

		assert.Equal(t, string(rr.Body.Bytes()), string(b))
	})

	t.Run("should return not found if incorrect path specified", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/notexistent.yaml", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		handler := NewStaticFileHandler("testdata/notexistent.yaml")

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}
