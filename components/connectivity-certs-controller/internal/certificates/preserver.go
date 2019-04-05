package certificates

import (
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"
	"github.com/pkg/errors"
)

const (
	clusterCertificateSecretKey = "crt"
	caCertificateSecretKey      = "ca.crt"
)

type Preserver interface {
	PreserveCertificates(certificates Certificates) error
}

type certificatePreserver struct {
	clusterCertSecretName string
	caCertSecretName      string
	secretsRepository     secrets.Repository
}

func NewCertificatePreserver(clusterCertSecretName string, caCertSecretName string, secretsRepository secrets.Repository) *certificatePreserver {
	return &certificatePreserver{
		clusterCertSecretName: clusterCertSecretName,
		caCertSecretName:      caCertSecretName,
		secretsRepository:     secretsRepository,
	}
}

func (cp *certificatePreserver) PreserveCertificates(certificates Certificates) error {
	err := cp.saveClusterCertificate(certificates.ClientCRT)
	if err != nil {
		return err
	}

	return cp.saveCACertificate(certificates.CaCRT)
}

func (cp *certificatePreserver) saveClusterCertificate(certificateChain []byte) error {
	clusterSecretData := map[string][]byte{
		clusterCertificateSecretKey: certificateChain,
	}

	err := cp.secretsRepository.UpsertWithMerge(cp.clusterCertSecretName, clusterSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to preserve client certificate in secret")
	}

	return nil
}

func (cp *certificatePreserver) saveCACertificate(caCertificate []byte) error {
	caSecretData := map[string][]byte{
		caCertificateSecretKey: caCertificate,
	}

	err := cp.secretsRepository.UpsertWithMerge(cp.caCertSecretName, caSecretData)
	if err != nil {
		return errors.Wrap(err, "Failed to preserve CA certificate in secret")
	}

	return nil
}
