package event_mesh_2_phases_prepare

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/kyma-project/kyma/common/resilient"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal"
	extsolutionhttp "github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/http"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/scenario"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type e2EEventMeshState struct {
	scenario.E2EState
	ServiceClassID string
	registryClient *testkit.RegistryClient
	dataStore      *testkit.DataStore
}

type PEMCertificate struct {
	Certificate [][]byte `json:"certificate"`
	PrivateKey  []byte `json:"private_key"`
}

func CertificatesToPEM(certificate tls.Certificate) (PEMCertificate, error) {
	rsaPK, ok := certificate.PrivateKey.(*rsa.PrivateKey)
	if ! ok {
		return PEMCertificate{}, fmt.Errorf("currently only rsa private keys are supported")
	}
	pem := PEMCertificate{
		Certificate: certificate.Certificate,
		PrivateKey: pem.EncodeToMemory(
			&pem.Block{
				Type: "RSA_PRIVATE_KEY",
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

// SetServiceClassID allows to set ServiceClassID so it can be shared between steps
func (s *e2EEventMeshState) SetServiceClassID(serviceID string) {
	s.ServiceClassID = serviceID

}

func (s *e2EEventMeshState) SetDataStore(dataStore *testkit.DataStore) {
	s.dataStore = dataStore
}

func (s *e2EEventMeshState) GetDataStore() *testkit.DataStore {
	return s.dataStore
}

// GetServiceClassID allows to get ServiceClassID so it can be shared between steps
func (s *e2EEventMeshState) GetServiceClassID() string {
	return s.ServiceClassID
}

// SetGatewayClientCerts allows to set application gateway client certificates so they can be used by later steps
func (s *e2EEventMeshState) SetGatewayClientCerts(certs []tls.Certificate) {
	metadataURL := fmt.Sprintf("https://gateway.%s/%s/v1/metadata/services", s.Domain, s.AppName)
	eventsUrl := fmt.Sprintf("https://gateway.%s/%s/events", s.Domain, s.AppName)

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(eventsUrl),
		cloudevents.WithBinaryEncoding(),
	)

	if err != nil {
		panic(err)
	}

	httpClient := internal.NewHTTPClient(s.SkipSSLVerify)
	httpClient.Transport.(*http.Transport).TLSClientConfig.Certificates = certs
	pemcerts := make([]PEMCertificate, len(certs))
	for i, cert := range certs {
		pemcert, err := CertificatesToPEM(cert);
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
	t.Client = httpClient
	client, err := cloudevents.NewClient(t)
	if err != nil {
		panic(err)
	}
	resilientEventClient := extsolutionhttp.NewWrappedCloudEventClient(client)

	resilientHTTPClient := resilient.WrapHttpClient(httpClient)
	s.registryClient = testkit.NewRegistryClient(metadataURL, resilientHTTPClient)
	s.EventSender = testkit.NewEventSender(nil, s.Domain, resilientEventClient)
}

func (s *e2EEventMeshState) GetRegistryClient() *testkit.RegistryClient {
	return s.registryClient
}
