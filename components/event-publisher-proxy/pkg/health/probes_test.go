package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealthSuccess(t *testing.T) {
	writer := httptest.NewRecorder()
	checkHealthHandler := CheckHealth(failHandler(http.StatusInternalServerError))

	requestLiveness := httptest.NewRequest(http.MethodGet, livenessURI, nil)
	checkHealthHandler.ServeHTTP(writer, requestLiveness)
	if got := writer.Result().StatusCode; got != http.StatusOK {
		t.Errorf("Handler must respond with status code %d to %s requests to path: '%s' but got code %d",
			http.StatusOK, http.MethodGet, livenessURI, got)
	}

	requestReadiness := httptest.NewRequest(http.MethodGet, readinessURI, nil)
	checkHealthHandler.ServeHTTP(writer, requestReadiness)
	if got := writer.Result().StatusCode; got != http.StatusOK {
		t.Errorf("Handler must respond with status code %d to %s requests to path: '%s' but got code %d",
			http.StatusOK, http.MethodGet, readinessURI, got)
	}
}

func TestCheckHealthFail(t *testing.T) {
	const (
		endpoint   = "/someEndpoint"
		statusCode = http.StatusInternalServerError
	)

	writer := httptest.NewRecorder()
	checkHealthHandler := CheckHealth(failHandler(statusCode))

	requestLiveness := httptest.NewRequest(http.MethodGet, endpoint, nil)
	checkHealthHandler.ServeHTTP(writer, requestLiveness)
	if got := writer.Result().StatusCode; got != statusCode {
		t.Errorf("Handler must respond with status code %d to %s requests to path: '%s' but got code %d",
			statusCode, http.MethodGet, endpoint, got)
	}
}

func failHandler(statusCode int) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(statusCode)
	})
}
