//go:build unit
// +build unit

package receiver

import (
	"net/http"
	"testing"

	testingutils "github.com/kyma-project/kyma/components/event-publisher-proxy/testing"
)

// a mocked http.Handler
type testHandler struct{}

func (h *testHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

var _ http.Handler = (*testHandler)(nil)

func TestNewHttpMessageReceiver(t *testing.T) {
	port := testingutils.GeneratePortOrDie()
	r := NewHTTPMessageReceiver(port)
	if r == nil {
		t.Fatalf("Could not create HTTPMessageReceiver")
	}
	if r.port != port {
		t.Errorf("Port should be: %d is: %d", port, r.port)
	}
}
