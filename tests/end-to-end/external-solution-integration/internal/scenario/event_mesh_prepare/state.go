package event_mesh_prepare

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/common/resilient"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type e2EEventMeshState struct {
	scenario.E2EState
	dataStore *testkit.DataStore
}

type PEMCertificate struct {
	Certificate [][]byte `json:"certificate"`
	PrivateKey  []byte   `json:"private_key"`
}

func CertificatesToPEM(certificate tls.Certificate) (PEMCertificate, error) {
	rsaPK, ok := certificate.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return PEMCertificate{}, fmt.Errorf("currently only rsa private keys are supported")
	}
	pem := PEMCertificate{
		Certificate: certificate.Certificate,
		PrivateKey: pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA_PRIVATE_KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(rsaPK),
			}),
	}
	return pem, nil
}

func PEMToCertificate(pemCert PEMCertificate) (tls.Certificate, error) {
	cert := tls.Certificate{}
	block, _ := pem.Decode(pemCert.PrivateKey)
	if block == nil {
		return cert, fmt.Errorf("failed parsing PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return cert, err
	}

	cert.Certificate = pemCert.Certificate
	cert.PrivateKey = privateKey
	return cert, nil
}

func (s *e2EEventMeshState) SetDataStore(dataStore *testkit.DataStore) {
	s.dataStore = dataStore
}

func (s *e2EEventMeshState) GetDataStore() *testkit.DataStore {
	return s.dataStore
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EEventMeshState) SetGatewayClientCerts(certs []tls.Certificate) {
	pemcerts := make([]PEMCertificate, len(certs))
	for i, cert := range certs {
		pemcert, err := CertificatesToPEM(cert)
		if err != nil {
			panic(err)
		}
		pemcerts[i] = pemcert
	}
	certJson, err := json.Marshal(pemcerts)
	if err != nil {
		panic(err)
	}
	if err := s.dataStore.Store(CertKey, string(certJson)); err != nil {
		panic(err)
	}
	if err := s.dataStore.Store(AppNameKey, s.AppName); err != nil {
		panic(err)
	}

	gatewayURL := fmt.Sprintf("https://%s.%s/%s/v1/metadata/services", s.GatewaySubdomain, s.Domain, s.AppName)
	httpClient := internal.NewHTTPClient(internal.WithSkipSSLVerification(s.SkipSSLVerify), internal.WithClientCertificates(certs))
	s.RegistryClient = testkit.NewRegistryClient(gatewayURL, resilient.WrapHttpClient(httpClient))
	s.EventSender = testkit.NewEventSender(httpClient, s.Domain, s.AppName)
}
