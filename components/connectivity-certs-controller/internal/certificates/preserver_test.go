package certificates

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"

	"github.com/stretchr/testify/require"
)

var (
	crtChain  = []byte("certificateChain")
	clientCRT = []byte("clientCertificate")
	caCRT     = []byte("caCertificate")
	clientKey = []byte("clientKey")
)

func TestCertificatePreserver_PreserveCertificates(t *testing.T) {

	certificates := Certificates{
		CRTChain:  crtChain,
		ClientCRT: clientCRT,
		CaCRT:     caCRT,
		ClientKey: clientKey,
	}

	t.Run("should preserve certificates", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   crtChain,
		}
		caSecretData := map[string][]byte{caCertificateSecretKey: caCRT}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertWithMerge", clusterCertSecretNamespaceName, clusterSecretData).Return(nil)
		secretsRepository.On("UpsertWithMerge", caCertSecretNamespaceName, caSecretData).Return(nil)

		certificatePreserver := NewCertificatePreserver(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.NoError(t, err)
		secretsRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to save cluster secret", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   crtChain,
		}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertWithMerge", clusterCertSecretNamespaceName, clusterSecretData).Return(errors.New("error"))

		certificatePreserver := NewCertificatePreserver(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.Error(t, err)
		secretsRepository.AssertExpectations(t)
	})

	t.Run("should return error when failed to save ca secret", func(t *testing.T) {
		// given
		clusterSecretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   crtChain,
		}
		caSecretData := map[string][]byte{caCertificateSecretKey: caCRT}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("UpsertWithMerge", clusterCertSecretNamespaceName, clusterSecretData).Return(nil)
		secretsRepository.On("UpsertWithMerge", caCertSecretNamespaceName, caSecretData).Return(errors.New("error"))

		certificatePreserver := NewCertificatePreserver(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := certificatePreserver.PreserveCertificates(certificates)

		// then
		require.Error(t, err)
		secretsRepository.AssertExpectations(t)
	})

}
