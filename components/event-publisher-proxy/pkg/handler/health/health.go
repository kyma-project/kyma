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

type Checker interface {
	ReadinessCheck(w http.ResponseWriter, r *http.Request)
	LivenessCheck(w http.ResponseWriter, r *http.Request)
}

// ConfigurableChecker represents a health checker.
type ConfigurableChecker struct {
	livenessCheck  http.HandlerFunc
	readinessCheck http.HandlerFunc
}

func (c ConfigurableChecker) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	c.readinessCheck(w, r)
}

func (c ConfigurableChecker) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	c.livenessCheck(w, r)
}

// CheckerOpt represents a health checker option.
type CheckerOpt func(*ConfigurableChecker)

// NewChecker returns a new instance of ConfigurableChecker initialized with the default liveness and readiness checks.
func NewChecker(opts ...CheckerOpt) *ConfigurableChecker {
	c := &ConfigurableChecker{
		livenessCheck:  DefaultCheck,
		readinessCheck: DefaultCheck,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
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

	return func(c *ConfigurableChecker) {
		c.livenessCheck = h
	}
}

// WithReadinessCheck returns CheckerOpt which sets the readiness check for the given http.HandlerFunc.
// It panics if the given http.HandlerFunc is nil.
func WithReadinessCheck(h http.HandlerFunc) CheckerOpt {
	if h == nil {
		panic("readiness handler is nil")
	}

	return func(c *ConfigurableChecker) {
		c.readinessCheck = h
	}
}
