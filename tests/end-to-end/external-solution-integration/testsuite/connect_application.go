package testsuite

import (
	"crypto/tls"
	"github.com/kyma-project/kyma/common/ingressgateway"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"net/http"
)

type ConnectApplication struct {
	connector *testkit.ConnectorClient
	state     ConnectApplicationState
}

type ConnectApplicationState interface {
	SetGatewayHttpClient(httpClient *http.Client)
}

func NewConnectApplication(connector *testkit.ConnectorClient, state ConnectApplicationState) *ConnectApplication {
	return &ConnectApplication{
		connector: connector,
		state:     state,
	}
}

func (s ConnectApplication) Name() string {
	return "Connect application"
}

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

	httpClient, err := ingressgateway.FromEnv().Client()
	if err != nil {
		return err
	}

	rawChain := make([][]byte, 0, len(chain))
	for _, cert := range chain {
		rawChain = append(rawChain, cert.Raw)
	}
	cert := tls.Certificate{Certificate: rawChain, PrivateKey: privateKey}
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = []tls.Certificate{cert}
	s.state.SetGatewayHttpClient(httpClient)
	return nil
}

func (s ConnectApplication) Cleanup() error {
	return s.connector.TokenRequestClient.DeleteTokenRequest()
}
