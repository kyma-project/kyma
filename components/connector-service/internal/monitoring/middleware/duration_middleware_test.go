package middleware

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/collector/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDurationMiddleware_Handle(t *testing.T) {

	t.Run("should observe status", func(t *testing.T) {
		// given
		path := "/v1/remoteenvironments/ec-default/tokens"

		collector := &mocks.Collector{}
		collector.On("AddObservation", path, http.MethodPost, mock.AnythingOfType("float64")).Return()

		codeMiddleware := NewDurationMiddleware(collector)

		router := mux.NewRouter()
		router.Use(codeMiddleware.Handle)
		router.Handle(path, getTestHandler())

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		var fullUrl bytes.Buffer
		fullUrl.WriteString(string(testServer.URL))
		fullUrl.WriteString(path)

		// when
		response, apperr := http.Post(fullUrl.String(), "application/json", nil)
		require.NoError(t, apperr)

		// then
		collector.AssertCalled(t, "AddObservation", path, http.MethodPost, mock.AnythingOfType("float64"))
		assert.Equal(t, testHandlerStatus, response.StatusCode)
	})

	t.Run("should fail observing status if route not matched", func(t *testing.T) {
		// given
		collector := &mocks.Collector{}

		codeMiddleware := NewCodeMiddleware(collector)

		req, err := http.NewRequest(http.MethodPost, "/some/path", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		handler := codeMiddleware.Handle(getTestHandler())
		handler.ServeHTTP(rr, req)

		// then
		collector.AssertNotCalled(t, "AddObservation")
	})
}
