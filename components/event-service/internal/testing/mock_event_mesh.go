package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	log "github.com/sirupsen/logrus"

	cloudevents "github.com/cloudevents/sdk-go"
)

type CloseFunction func()

// StartMockEventMeshServer mocks the event mesh with the given HTTP response code and returns its URL as a string.
func StartMockEventMeshServer(t *testing.T, respCode int) (string, CloseFunction) {
	t.Helper()
	log.Info("Initialising the mock server...")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(respCode)
		resp := &cloudevents.EventResponse{
			Status: respCode,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to write response")
		}
	}))
	return srv.URL, func() {
		log.Info("Closing the mock server...")
		srv.Close()
	}
}
