package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"

	"github.com/pkg/errors"
)

type Provider interface {// TODO - consider moving to the same component as preserver
	GetClientCredentials() (*rsa.PrivateKey, *x509.Certificate, error)
}

type certificateProvider struct {
	clusterCertSecretName string
	//caCRTSecretName       string
	secretsRepository secrets.Repository
}

func NewCertificateProvider(clusterCertSecretName string, secretsRepository secrets.Repository) Provider {
	return &certificateProvider{
		secretsRepository:     secretsRepository,
		clusterCertSecretName: clusterCertSecretName,
	}
}

func (cp *certificateProvider) GetClientCredentials() (*rsa.PrivateKey, *x509.Certificate, error) {
	secretData, err := cp.secretsRepository.Get(cp.clusterCertSecretName)
	if err != nil {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("Failed to read %s secret with certificates", cp.clusterCertSecretName))
	}

	clientCert, err := getClientCertificate(secretData[clusterCertificateSecretKey])
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to read client certificate")
	}

	clientKey, err := getClientPrivateKey(secretData[clusterKeySecretKey])
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to read client key")
	}

	return clientKey, clientCert, nil
}

func getClientCertificate(clusterCert []byte) (*x509.Certificate, error) {
	if clusterCert == nil {
		return nil, errors.New("Cluster certificate data not found")
	}

	block, _ := pem.Decode(clusterCert)
	if block == nil {
		return nil, errors.New("Failed to decode client certificate pem")
	}

	return x509.ParseCertificate(block.Bytes)
}

func getClientPrivateKey(clusterKey []byte) (*rsa.PrivateKey, error) {
	if clusterKey == nil {
		return nil, errors.New("Client key data not found")
	}

	block, _ := pem.Decode(clusterKey)
	if block == nil {
		return nil, errors.New("Failed to decode client key pem")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
