package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"

	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
	clusterPrivateKey, err := cp.provideClusterPrivateKey()
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

// TODO - always create new key
func (cp *csrProvider) provideClusterPrivateKey() (*rsa.PrivateKey, error) {
	secret, err := cp.secretRepository.Get(cp.clusterCertSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return cp.createClusterKeySecret()
		}
		return nil, errors.Wrapf(err, fmt.Sprintf("Failed to read cluster %s secret", cp.clusterCertSecretName))
	}

	block, _ := pem.Decode(secret[clusterKeySecretKey])
	if block == nil {
		return cp.createClusterKeySecret()
	}

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error while parsing private key")
	}

	return privateKey.(*rsa.PrivateKey), nil
}

func (cp *csrProvider) createClusterKeySecret() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}

	secretData := map[string][]byte{
		clusterKeySecretKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	}

	err = cp.secretRepository.UpsertWithReplace(cp.clusterCertSecretName, secretData)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to override cluster key secret")
	}

	return key, nil
}
