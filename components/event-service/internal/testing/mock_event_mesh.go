package testing

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go"
)

// MockEventMesh mocks the event mesh and returns its URL as a string.
func MockEventMesh(t *testing.T) *string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &cloudevents.EventResponse{Status: http.StatusOK}
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(resp); err != nil {
			t.Fatalf("failed to write response")
		}
	}))

	if srv == nil {
		t.Fatalf("failed to start HTTP server")
		return nil
	}

	return &srv.URL
}
