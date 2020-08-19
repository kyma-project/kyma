package certificates

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets/mocks"
	"github.com/stretchr/testify/require"
)

const (
	clusterCertSecretName      = "cluster-certificate"
	clusterCertSecretNamespace = "kyma-integration"
	caCertSecretName           = "ca-cert"
	caCertSecretNamespace      = "istio-system"
)

var (
	clusterCertSecretNamespaceName = types.NamespacedName{
		Name:      clusterCertSecretName,
		Namespace: clusterCertSecretNamespace,
	}
	caCertSecretNamespaceName = types.NamespacedName{
		Name:      caCertSecretName,
		Namespace: caCertSecretNamespace,
	}
)

func TestCertificatePreserver_PreserveCertificates(t *testing.T) {

	pemCredentials := PemEncodedCredentials{
		ClientKey:         clientKey,
		CertificateChain:  crtChain,
		ClientCertificate: clientCRT,
		CACertificates:    caCRT,
	}

	credentials, err := pemCredentials.AsCredentials()
	require.NoError(t, err)

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

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := credentialsManager.PreserveCredentials(credentials)

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

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := credentialsManager.PreserveCredentials(credentials)

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

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		err := credentialsManager.PreserveCredentials(credentials)

		// then
		require.Error(t, err)
		secretsRepository.AssertExpectations(t)
	})
}

func TestCertificateProvider_GetClientCredentials(t *testing.T) {

	pemCredentials := PemEncodedCredentials{
		ClientKey:         clientKey,
		CertificateChain:  crtChain,
		ClientCertificate: clientCRT,
		CACertificates:    caCRT,
	}

	t.Run("should get client credentials", func(t *testing.T) {
		// given
		expectedCreds, err := pemCredentials.AsClientCredentials()
		require.NoError(t, err)

		secretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   crtChain,
		}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		clientCreds, err := credentialsManager.GetClientCredentials()

		// then
		require.NoError(t, err)
		require.NotNil(t, clientCreds)
		assert.Equal(t, expectedCreds, clientCreds)
	})

	t.Run("should return error when failed to read secret", func(t *testing.T) {
		// given
		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(nil, errors.New("error"))

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		_, err := credentialsManager.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when no data in secret", func(t *testing.T) {
		// given
		secretData := map[string][]byte{}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		_, err := credentialsManager.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode cert", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: []byte("invalid pem"),
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   crtChain,
		}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		_, err := credentialsManager.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode certificate chain", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         clientKey,
			certificateChainSecretKey:   []byte("invalid pem"),
		}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		_, err := credentialsManager.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode client key", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: clientCRT,
			clusterKeySecretKey:         []byte("invalid pem"),
			certificateChainSecretKey:   crtChain,
		}

		secretsRepository := &mocks.Repository{}
		secretsRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		credentialsManager := NewCredentialsManager(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretsRepository)

		// when
		_, err := credentialsManager.GetClientCredentials()

		// then
		require.Error(t, err)
	})
}
