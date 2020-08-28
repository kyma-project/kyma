package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"

	"github.com/pkg/errors"
)

const (
	rsaKeySize = 4096
)

//go:generate mockery --name=CSRProvider
type CSRProvider interface {
	CreateCSR(subject pkix.Name) (string, *rsa.PrivateKey, error)
}

type csrProvider struct {
}

func NewCSRProvider() CSRProvider {
	return &csrProvider{}
}

// CreateCSR generates private key and returns it along with base 64 encoded CSR
func (cp *csrProvider) CreateCSR(subject pkix.Name) (string, *rsa.PrivateKey, error) {
	clusterPrivateKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return "", nil, err
	}

	csr, err := createCSR(subject, clusterPrivateKey)
	if err != nil {
		return "", nil, err
	}

	return base64.StdEncoding.EncodeToString(csr), clusterPrivateKey, nil
}

func createCSR(subject pkix.Name, key *rsa.PrivateKey) ([]byte, error) {
	csrTemplate := x509.CertificateRequest{
		Subject: subject,
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create cluster CSR")
	}

	pemEncodedCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})

	return pemEncodedCSR, nil
}
