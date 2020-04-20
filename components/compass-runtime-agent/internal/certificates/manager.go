package certificates

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	clusterCertificateSecretKey = "crt"
	clusterKeySecretKey         = "key"
	certificateChainSecretKey   = "crtChain"

	caCertificateSecretKey = "cacert"
)

//go:generate mockery -name=Manager
type Manager interface {
	GetClientCredentials() (ClientCredentials, error)
	PreserveCredentials(Credentials) error
}

func NewCredentialsManager(clusterCertificateSecretName, caCertSecretName types.NamespacedName, secretsRepository secrets.Repository) *credentialsManager {
	return &credentialsManager{
		caCertSecretName:              caCertSecretName,
		clusterCertificatesSecretName: clusterCertificateSecretName,
		secretsRepository:             secretsRepository,
	}
}

type credentialsManager struct {
	caCertSecretName              types.NamespacedName
	clusterCertificatesSecretName types.NamespacedName
	secretsRepository             secrets.Repository
}

func (cm *credentialsManager) GetClientCredentials() (ClientCredentials, error) {
	secretData, err := cm.secretsRepository.Get(cm.clusterCertificatesSecretName)
	if err != nil {
		return ClientCredentials{}, errors.Wrap(err, fmt.Sprintf("Failed to read %s secret with certificates", cm.clusterCertificatesSecretName))
	}

	pemCredentials := PemEncodedCredentials{
		ClientKey:         secretData[clusterKeySecretKey],
		CertificateChain:  secretData[certificateChainSecretKey],
		ClientCertificate: secretData[clusterCertificateSecretKey],
	}

	return pemCredentials.AsClientCredentials()
}

func (cm *credentialsManager) PreserveCredentials(credentials Credentials) error {
	pemCredentials := credentials.AsPemEncoded()

	err := cm.saveClusterCertificateAndKey(pemCredentials.ClientKey, pemCredentials.ClientCertificate, pemCredentials.CertificateChain)
	if err != nil {
		return err
	}

	return cm.saveCACertificate(pemCredentials.CACertificates)
}

func (cm *credentialsManager) saveClusterCertificateAndKey(clientKey, clientCert, certificateChain []byte) error {
	clusterSecretData := map[string][]byte{
		clusterCertificateSecretKey: clientCert,
		clusterKeySecretKey:         clientKey,
		certificateChainSecretKey:   certificateChain,
	}

	err := cm.secretsRepository.UpsertWithMerge(cm.clusterCertificatesSecretName, clusterSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to preserve client certificate and key in secret")
	}

	return nil
}

func (cm *credentialsManager) saveCACertificate(caCertificate []byte) error {
	caSecretData := map[string][]byte{
		caCertificateSecretKey: caCertificate,
	}

	err := cm.secretsRepository.UpsertWithMerge(cm.caCertSecretName, caSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to preserve CA certificate in secret")
	}

	return nil
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
