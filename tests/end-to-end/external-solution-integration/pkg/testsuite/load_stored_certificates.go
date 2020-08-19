package testsuite

import (
	"crypto/tls"
	"encoding/json"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

type LoadStoredCertificates struct {
	ds    *testkit.DataStore
	state LoadStoredCertificatesState
}

type LoadStoredCertificatesState interface {
	SetGatewayClientCerts(certs []tls.Certificate)
	SetApplicationName(string)
}

//  NewLoadStoredCertificates loads previously stored certificates
func NewLoadStoredCertificates(ds *testkit.DataStore, state LoadStoredCertificatesState) *LoadStoredCertificates {
	return &LoadStoredCertificates{
		ds:    ds,
		state: state,
	}
}

// Name returns name name of the step
func (s LoadStoredCertificates) Name() string {
	return "Load Stored Certificates"
}

// Run executes the step
func (s LoadStoredCertificates) Run() error {
	appName, err := s.ds.Load(AppNameKey)
	if err != nil {
		return err
	}
	certsJson, err := s.ds.Load(CertKey)
	if err != nil {
		return err
	}
	var pemcerts []PEMCertificate
	if err := json.Unmarshal([]byte(certsJson), &pemcerts); err != nil {
		return err
	}

	certs := make([]tls.Certificate, len(pemcerts))
	for i, pemcert := range pemcerts {
		cert, err := PEMToCertificate(pemcert)
		if err != nil {
			return err
		}
		certs[i] = cert
	}
	s.state.SetApplicationName(appName)
	s.state.SetGatewayClientCerts(certs)
	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s LoadStoredCertificates) Cleanup() error {
	return nil
}
