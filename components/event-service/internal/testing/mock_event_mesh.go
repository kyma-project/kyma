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

// MockEventMesh mocks the event mesh and returns its URL as a string.
func MockEventMesh(t *testing.T) (string, CloseFunction) {
	t.Helper()
	log.Info("Initialising the mock server...")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &cloudevents.EventResponse{Status: http.StatusOK}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			t.Fatalf("failed to write response")
		}
	}))
	return srv.URL, func() {
		log.Info("Closing the mock server...")
		srv.Close()
	}
}
