package nats

import (
	"net/http"

	"github.com/nats-io/nats.go"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"
)

const healthCheckName = "health-check"

// ReadinessCheck returns an instance of http.HandlerFunc that checks the readiness of the given NATS Handler.
// It checks the NATS server connection status and reports 2XX if connected, otherwise reports 5XX.
// It panics if the given NATS Handler is nil.
func ReadinessCheck(h *Handler) http.HandlerFunc {
	if h == nil {
		panic("readiness handler is nil")
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		s := h.Sender
		if status := s.ConnectionStatus(); status != nats.CONNECTED {
			h.Logger.WithContext().Named(healthCheckName).With("connection-status", status).Info("Disconnected from NATS server")
			w.WriteHeader(health.StatusCodeNotHealthy)
			return
		}

		w.WriteHeader(health.StatusCodeHealthy)
	}
}
