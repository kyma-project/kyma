package middleware

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/connector-service/internal/middleware/metrics/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
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
		path := "/v1/remoteenvironments/ec-default/tokens"

		collector := &mocks.Collector{}
		collector.On("AddObservation", path, http.MethodPost, float64(testHandlerStatus)).Return()

		codeMiddleware := NewCodeMiddleware(collector)

		router := mux.NewRouter()
		router.Handle(path, codeMiddleware.Handle(getTestHandler()))

		testServer := httptest.NewServer(router)
		defer testServer.Close()

		var fullUrl bytes.Buffer
		fullUrl.WriteString(string(testServer.URL))
		fullUrl.WriteString(path)

		// when
		response, apperr := http.Post(fullUrl.String(), "application/json", nil)
		require.NoError(t, apperr)

		// then
		collector.AssertCalled(t, "AddObservation", path, http.MethodPost, float64(testHandlerStatus))
		assert.Equal(t, testHandlerStatus, response.StatusCode)
	})
}
