package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"

	"github.com/pkg/errors"
)

type Provider interface {
	GetClientCredentials() (*rsa.PrivateKey, *x509.Certificate, error)
	GetCACertificates() ([]*x509.Certificate, error)
}

type certificateProvider struct {
	clusterCertSecretName string
	caCertSecretName      string
	secretsRepository     secrets.Repository
}

func NewCertificateProvider(clusterCertSecretName string, caCertSecretName string, secretsRepository secrets.Repository) Provider {
	return &certificateProvider{
		secretsRepository:     secretsRepository,
		caCertSecretName:      caCertSecretName,
		clusterCertSecretName: clusterCertSecretName,
	}
}

func (cp *certificateProvider) GetCACertificates() ([]*x509.Certificate, error) {
	secretData, err := cp.secretsRepository.Get(cp.caCertSecretName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to read %s secret with certificates", cp.clusterCertSecretName))
	}

	caCerts, err := decodeCertificates(secretData[caCertificateSecretKey])
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read client certificate")
	}

	return caCerts, nil
}

func (cp *certificateProvider) GetClientCredentials() (*rsa.PrivateKey, *x509.Certificate, error) {
	secretData, err := cp.secretsRepository.Get(cp.clusterCertSecretName)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Failed to read %s secret with certificates", cp.clusterCertSecretName))
	}

	clientCert, err := decodeCertificate(secretData[clusterCertificateSecretKey])
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to read client certificate")
	}

	clientKey, err := getClientPrivateKey(secretData[clusterKeySecretKey])
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to read client key")
	}

	return clientKey, clientCert, nil
}

func decodeCertificate(certificate []byte) (*x509.Certificate, error) {
	certs, err := decodeCertificates(certificate)
	if err != nil {
		return nil, err
	}

	return certs[0], nil
}

func decodeCertificates(certificate []byte) ([]*x509.Certificate, error) {
	if certificate == nil {
		return nil, errors.New("Certificate data is empty")
	}

	var certificates []*x509.Certificate

	for block, rest := pem.Decode(certificate); block != nil && rest != nil; {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to decode one of the pem blocks")
		}

		certificates = append(certificates, cert)

		block, rest = pem.Decode(rest)
	}

	if len(certificates) == 0 {
		return nil, errors.New("No certificates found in the pem block")
	}

	return certificates, nil
}

func getClientPrivateKey(clusterKey []byte) (*rsa.PrivateKey, error) {
	if clusterKey == nil {
		return nil, errors.New("Private key data is empty")
	}

	block, _ := pem.Decode(clusterKey)
	if block == nil {
		return nil, errors.New("Failed to decode client key pem")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
