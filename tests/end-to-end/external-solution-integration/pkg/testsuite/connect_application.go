package testsuite

import (
	"crypto/tls"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

// ConnectApplication is a step which connects application with client certificates and saves connected httpClient in the state
type ConnectApplication struct {
	connector *testkit.ConnectorClient
	state     ConnectApplicationState
}

// ConnectApplicationState allows ConnectApplication to save connected http.Client for further use by other steps
type ConnectApplicationState interface {
	SetGatewayClientCerts(certs []tls.Certificate)
}

// NewConnectApplication returns new ConnectApplication
func NewConnectApplication(connector *testkit.ConnectorClient, state ConnectApplicationState) *ConnectApplication {
	return &ConnectApplication{
		connector: connector,
		state:     state,
	}
}

// Name returns name name of the step
func (s ConnectApplication) Name() string {
	return "Connect application"
}

// Run executes the step
func (s ConnectApplication) Run() error {
	infoURL, err := s.connector.GetToken()
	if err != nil {
		return err
	}

	certInfo, err := s.connector.GetInfo(infoURL)
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

	chain, err := s.connector.GetCertificate(certInfo.CertUrl, csr)
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
func (s ConnectApplication) Cleanup() error {
	return s.connector.TokenRequestClient.DeleteTokenRequest()
}
