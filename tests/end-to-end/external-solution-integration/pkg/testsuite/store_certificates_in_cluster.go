package testsuite

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/testkit"
)

const (
	CertKey    = "certificates"
	AppNameKey = "appname"
)

type PEMCertificate struct {
	Certificate [][]byte `json:"certificate"`
	PrivateKey  []byte   `json:"private_key"`
}

type StoreCertificatesInCluster struct {
	getCertificates func() []tls.Certificate
	ds              *testkit.DataStore
	appName         string
}

func NewStoreCertificatesInCluster(ds *testkit.DataStore, applicationName string, get func() []tls.Certificate) *StoreCertificatesInCluster {
	return &StoreCertificatesInCluster{
		ds:              ds,
		getCertificates: get,
		appName:         applicationName,
	}
}

// Name returns name name of the step
func (s StoreCertificatesInCluster) Name() string {
	return "Store Certificates in Cluster"
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

// Run executes the step
func (s StoreCertificatesInCluster) Run() error {
	certificates := s.getCertificates()
	pemcerts := make([]PEMCertificate, len(certificates))
	for i, cert := range certificates {
		pemcert, err := CertificatesToPEM(cert)
		if err != nil {
			return err
		}
		pemcerts[i] = pemcert
	}
	certJson, err := json.Marshal(pemcerts)
	if err != nil {
		return err
	}
	if err := s.ds.Store(CertKey, string(certJson)); err != nil {
		return err
	}
	if err := s.ds.Store(AppNameKey, s.appName); err != nil {
		return err
	}

	return nil
}

// Cleanup removes all resources that may possibly created by the step
func (s StoreCertificatesInCluster) Cleanup() error {
	return s.ds.Destroy()
}
