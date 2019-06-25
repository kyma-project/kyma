package certificates

import (
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	clusterCertificateSecretKey = "crt"
	clusterKeySecretKey         = "key"

	caCertificateSecretKey = "cacert"
)

type Preserver interface {
	PreserveCertificates(certificates Certificates) error
}

type certificatePreserver struct {
	clusterCertSecretName types.NamespacedName
	caCertSecretName      types.NamespacedName
	secretsRepository     secrets.Repository
}

func NewCertificatePreserver(clusterCertSecret types.NamespacedName, caCertSecret types.NamespacedName, secretsRepository secrets.Repository) *certificatePreserver {
	return &certificatePreserver{
		clusterCertSecretName: clusterCertSecret,
		caCertSecretName:      caCertSecret,
		secretsRepository:     secretsRepository,
	}
}

func (cp *certificatePreserver) PreserveCertificates(certificates Certificates) error {
	err := cp.saveClusterCertificateAndKey(certificates.ClientKey, certificates.ClientCRT)
	if err != nil {
		return err
	}

	return cp.saveCACertificate(certificates.CaCRT)
}

func (cp *certificatePreserver) saveClusterCertificateAndKey(clientKey, certificateChain []byte) error {
	clusterSecretData := map[string][]byte{
		clusterCertificateSecretKey: certificateChain,
		clusterKeySecretKey:         clientKey,
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
