package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"

	"github.com/pkg/errors"
)

const (
	rsaKeySize          = 4096
	clusterKeySecretKey = "key"
)

type CSRProvider interface {
	CreateCSR(subject pkix.Name) (string, error)
}

type csrProvider struct {
	clusterCertSecretName string
	caCRTSecretName       string
	secretRepository      secrets.Repository
}

func NewCSRProvider(clusterCertSecret, caCRTSecret string, secretRepository secrets.Repository) CSRProvider {
	return &csrProvider{
		clusterCertSecretName: clusterCertSecret,
		caCRTSecretName:       caCRTSecret,
		secretRepository:      secretRepository,
	}
}

func (cp *csrProvider) CreateCSR(subject pkix.Name) (string, error) {
	clusterPrivateKey, err := cp.createClusterKeySecret()
	if err != nil {
		return "", err
	}

	csr, err := createCSR(subject, clusterPrivateKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(csr), nil
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

func (cp *csrProvider) createClusterKeySecret() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}

	secretData := map[string][]byte{
		clusterKeySecretKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	}

	err = cp.secretRepository.UpsertWithMerge(cp.clusterCertSecretName, secretData)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to override cluster key secret")
	}

	return key, nil
}
