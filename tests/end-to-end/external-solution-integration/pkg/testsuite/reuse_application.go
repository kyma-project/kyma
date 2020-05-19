package testsuite

import (
	"crypto/tls"
)

// ConnectApplication is a step which connects application with client certificates and saves connected httpClient in the state
type ReuseApplication struct {
	state ReuseApplicationState
}

// ConnectApplicationState allows ConnectApplication to save connected http.Client for further use by other steps
type ReuseApplicationState interface {
	SetGatewayClientCerts(certs []tls.Certificate)
}

// NewConnectApplication returns new ConnectApplication
func NewReuseApplication(state ConnectApplicationState) *ReuseApplication {
	return &ReuseApplication{
		state: state,
	}
}

// Name returns name name of the step
func (s ReuseApplication) Name() string {
	return "Reuse application"
}

// Run executes the step
func (s ReuseApplication) Run() error {
	s.state.SetGatewayClientCerts([]tls.Certificate{})
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s ReuseApplication) Cleanup() error {
	return nil
}
