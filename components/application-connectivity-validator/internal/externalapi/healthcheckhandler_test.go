package externalapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheckHandler_HandleRequest(t *testing.T) {
	t.Run("should always respond with 200 status code", func(t *testing.T) {
		// given
		req, err := http.NewRequest(http.MethodGet, "/v1/health", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		handler := NewHealthCheckHandler()

		// when
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
