package certificates

import (
	"errors"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	crtChain  = []byte("certificateChain")
	clientCRT = []byte("clientCertificate")
	caCRT     = []byte("caCertificate")
)

func TestCertificatePreserver_PreserveCertificates(t *testing.T) {

	clusterCertSecretName := "cluster-secret-name"
	caSecretName := "ca-secret-name"

	certificates := Certificates{
		CRTChain:  crtChain,
		ClientCRT: clientCRT,
		CaCRT:     caCRT,
	}

	t.Run("should preserve certificates", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{clusterCertificateSecretKey: crtChain}
		caSecretData := map[string][]byte{caCertificateSecretKey: caCRT}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertData", clusterCertSecretName, clusterSecretData).Return(nil)
		secretsRepository.On("UpsertData", caSecretName, caSecretData).Return(nil)

		certificatePreserver := NewCertificatePreserver(clusterCertSecretName, caSecretName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.NoError(t, err)
		secretsRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to save cluster secret", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{clusterCertificateSecretKey: crtChain}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertData", clusterCertSecretName, clusterSecretData).Return(errors.New("error"))

		certificatePreserver := NewCertificatePreserver(clusterCertSecretName, caSecretName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.Error(t, err)
		secretsRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to save ca secret", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{clusterCertificateSecretKey: crtChain}
		caSecretData := map[string][]byte{caCertificateSecretKey: caCRT}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertData", clusterCertSecretName, clusterSecretData).Return(nil)
		secretsRepository.On("UpsertData", caSecretName, caSecretData).Return(errors.New("error"))

		certificatePreserver := NewCertificatePreserver(clusterCertSecretName, caSecretName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.Error(t, err)
		secretsRepository.AssertExpectations(t)
	})

}
