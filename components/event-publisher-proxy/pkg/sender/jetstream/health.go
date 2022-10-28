package jetstream

import (
	"net/http"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
)

// ReadinessCheck returns an instance of http.HandlerFunc that checks the readiness of the given NATS Handler.
// It checks the NATS server connection status and reports 2XX if connected, otherwise reports 5XX.
// It panics if the given NATS Handler is nil.
func (s *Sender) ReadinessCheck(w http.ResponseWriter, _ *http.Request) {
	if status := s.ConnectionStatus(); status != nats.CONNECTED {
		s.namedLogger().Error("Readiness check failed: not connected to nats server")
		w.WriteHeader(health.StatusCodeNotHealthy)
		return
	}
	w.WriteHeader(health.StatusCodeHealthy)
}

func (s *Sender) LivenessCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(health.StatusCodeHealthy)
}
