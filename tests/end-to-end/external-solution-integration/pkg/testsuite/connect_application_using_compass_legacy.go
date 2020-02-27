package testsuite

import (
	"crypto/tls"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// ConnectApplicationUsingCompass is a step which connects application with client certificates and saves connected httpClient in the state
type ConnectApplicationUsingCompassLegacy struct {
	connector       *testkit.CompassConnectorClient
	legacyConnector *testkit.ConnectorClient
	director        *testkit.CompassDirectorClient
	state           ConnectApplicationUsingCompassLegacyState
}

// ConnectApplicationUsingCompassState allows ConnectApplicationUsingCompass to save connected http.Client for further use by other steps
type ConnectApplicationUsingCompassLegacyState interface {
	SetGatewayClientCerts(certs []tls.Certificate)
	GetCompassAppID() string
}

// NewConnectApplicationUsingCompass returns new ConnectApplicationUsingCompass
func NewRegisterServiceUsingConnectivityAdapter(connector *testkit.CompassConnectorClient, legacyConnector *testkit.ConnectorClient,
	director *testkit.CompassDirectorClient, state ConnectApplicationUsingCompassLegacyState) *ConnectApplicationUsingCompassLegacy {
	return &ConnectApplicationUsingCompassLegacy{
		connector:       connector,
		legacyConnector: legacyConnector,
		director:        director,
		state:           state,
	}
}

// Name returns name name of the step
func (s ConnectApplicationUsingCompassLegacy) Name() string {
	return "Connect application using Compass"
}

// Run executes the step
func (s ConnectApplicationUsingCompassLegacy) Run() error {
	oneTimeToken, err := s.director.GetOneTimeTokenForApplication(s.state.GetCompassAppID())
	if err != nil {
		return err
	}

	certInfo, err := s.legacyConnector.GetInfo(oneTimeToken.LegacyConnectorURL)
	if err != nil {
		return err
	}

	privateKey, err := testkit.CreateKey()
	if err != nil {
		return err
	}

	csr, err := testkit.CreateCSR(certInfo.Certificate.Subject, privateKey)
	if err != nil {
		return err
	}

	chain, err := s.legacyConnector.GetCertificate(certInfo.CertUrl, csr)
	if err != nil {
		return err
	}

	rawChain := make([][]byte, 0, len(chain))
	for _, cert := range chain {
		rawChain = append(rawChain, cert.Raw)
	}
	cert := tls.Certificate{Certificate: rawChain, PrivateKey: privateKey}

	s.state.SetGatewayClientCerts([]tls.Certificate{cert})
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s ConnectApplicationUsingCompassLegacy) Cleanup() error {
	return nil
}
