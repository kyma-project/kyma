package health

import (
	"net/http"
)

const (
	// LivenessURI is the endpoint URI used for liveness check.
	LivenessURI = "/healthz"

	// ReadinessURI is the endpoint URI used for readiness check.
	ReadinessURI = "/readyz"

	// StatusCodeHealthy is the status code which indicates a healthy state.
	StatusCodeHealthy = http.StatusOK

	// StatusCodeNotHealthy is the status code which indicates a not healthy state.
	StatusCodeNotHealthy = http.StatusInternalServerError
)

// Checker represents a health checker.
type Checker struct {
	livenessCheck  http.HandlerFunc
	readinessCheck http.HandlerFunc
}

// CheckerOpt represents a health checker option.
type CheckerOpt func(*Checker)

// NewChecker returns a new instance of Checker initialized with the default liveness and readiness checks.
func NewChecker(opts ...CheckerOpt) *Checker {
	c := &Checker{
		livenessCheck:  DefaultCheck,
		readinessCheck: DefaultCheck,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Check does the necessary health checks (if applicable) before serving HTTP requests for the given http.Handler.
func (c *Checker) Check(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.livenessCheck != nil && r.RequestURI == LivenessURI {
			c.livenessCheck(w, r)
			return
		}

		if c.readinessCheck != nil && r.RequestURI == ReadinessURI {
			c.readinessCheck(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// DefaultCheck always writes a 2XX status code for the given http.ResponseWriter.
func DefaultCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(StatusCodeHealthy)
}

// WithLivenessCheck returns CheckerOpt which sets the liveness check for the given http.HandlerFunc.
// It panics if the given http.HandlerFunc is nil.
func WithLivenessCheck(h http.HandlerFunc) CheckerOpt {
	if h == nil {
		panic("liveness handler is nil")
	}

	return func(c *Checker) {
		c.livenessCheck = h
	}
}

// WithReadinessCheck returns CheckerOpt which sets the readiness check for the given http.HandlerFunc.
// It panics if the given http.HandlerFunc is nil.
func WithReadinessCheck(h http.HandlerFunc) CheckerOpt {
	if h == nil {
		panic("readiness handler is nil")
	}

	return func(c *Checker) {
		c.readinessCheck = h
	}
}
