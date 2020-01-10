package testsuite

import (
	"crypto/tls"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// ConnectApplicationUsingCompass is a step which connects application with client certificates and saves connected httpClient in the state
type ConnectApplicationUsingCompass struct {
	connector *testkit.CompassConnectorClient
	director  *testkit.CompassDirectorClient
	state     ConnectApplicationUsingCompassState
}

// ConnectApplicationUsingCompassState allows ConnectApplicationUsingCompass to save connected http.Client for further use by other steps
type ConnectApplicationUsingCompassState interface {
	SetGatewayClientCerts(certs []tls.Certificate)
	GetCompassAppID() string
}

// NewConnectApplicationUsingCompass returns new ConnectApplicationUsingCompass
func NewConnectApplicationUsingCompass(connector *testkit.CompassConnectorClient, director *testkit.CompassDirectorClient, state ConnectApplicationUsingCompassState) *ConnectApplicationUsingCompass {
	return &ConnectApplicationUsingCompass{
		connector: connector,
		director:  director,
		state:     state,
	}
}

// Name returns name name of the step
func (s ConnectApplicationUsingCompass) Name() string {
	return "Connect application using Compass"
}

// Run executes the step
func (s ConnectApplicationUsingCompass) Run() error {
	oneTimeToken, err := s.director.GetOneTimeTokenForApplication(s.state.GetCompassAppID())
	if err != nil {
		return err
	}

	certificate, err := s.connector.GenerateCertificateForToken(oneTimeToken.Token, oneTimeToken.ConnectorURL)
	if err != nil {
		return err
	}

	s.state.SetGatewayClientCerts([]tls.Certificate{certificate})

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s ConnectApplicationUsingCompass) Cleanup() error {
	return nil
}
