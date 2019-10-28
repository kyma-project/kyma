package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring/collector/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHandlerStatus = http.StatusOK

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(testHandlerStatus)
}

func getTestHandler() http.HandlerFunc {
	return http.HandlerFunc(serveHTTP)
}

func TestCodeMiddleware_Handle(t *testing.T) {

	t.Run("should observe status", func(t *testing.T) {
		// given
		path := "/v1/applications/ec-default/tokens"

		collector := &mocks.Collector{}
		collector.On("AddObservation", float64(1), path, "200", http.MethodPost).Return()

		codeMiddleware := NewCodeMiddleware(collector)

		router := mux.NewRouter()
		router.Use(codeMiddleware.Handle)
		router.Handle(path, getTestHandler())

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		fullUrl := fmt.Sprintf("%s%s", testServer.URL, path)

		// when
		response, apperr := http.Post(fullUrl, "application/json", nil)
		require.NoError(t, apperr)

		// then
		collector.AssertCalled(t, "AddObservation", float64(1), path, "200", http.MethodPost)
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
