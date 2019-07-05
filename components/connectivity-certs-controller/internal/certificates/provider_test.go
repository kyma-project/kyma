package certificates

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets/mocks"
)

const (
	clusterCertSecretName      = "cluster-certificate"
	clusterCertSecretNamespace = "kyma-integration"
	caCertSecretName           = "ca-cert"
	caCertSecretNamespace      = "istio-system"

	pemCertificate = `-----BEGIN CERTIFICATE-----
MIIEcDCCA1igAwIBAgIBAjANBgkqhkiG9w0BAQsFADBfMQ8wDQYDVQQLDAZDNGNv
cmUxDDAKBgNVBAoMA1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwHV2Fs
ZG9yZjELMAkGA1UEBhMCREUxDTALBgNVBAMMBEt5bWEwHhcNMTkwMjEyMTIwNDM4
WhcNMjAwMjEyMTIwNDM4WjBvMQswCQYDVQQGEwJERTEQMA4GA1UECBMHV2FsZG9y
ZjEQMA4GA1UEBxMHV2FsZG9yZjEVMBMGA1UEChMMT3JnYW5pemF0aW9uMRAwDgYD
VQQLEwdPcmdVbml0MRMwEQYDVQQDEwplYy1kZWZhdWx0MIICIjANBgkqhkiG9w0B
AQEFAAOCAg8AMIICCgKCAgEAq4TMGWybHjvhF9RSSZJG8zfp73RJlRhPj3e4EJgr
z0Ai/PmHa8WeK2eVAfsoky9UVE+1t+cn+5trejIzSMf1R8AKpgcvTDkO9RPRLqKm
3u8CjvOrJn0tKSk1Jf9kdSY9xQzd4SOnzSjhbL44MV3zpQ5qdlA2vIvCVNK0SwwS
xqVQbI4FEmbOjtvQKpxnkov2fjviAqcRd+PhR5lNnCGKGtNoVi9lEnXRCdfQsn5C
V67+y8MQkdswBdGrAdSgjTwLvI9kp7eiHAFyRJsLxcFT4VwDFJ2LrEm1hcs/5mDp
UnJWig3g6yVW9ME9oxy7F2etxUsJ4WRFfxixma0hW1AQk1LZduPVmRIHroIDKJA3
79C2gk4b75usaH3gL9s7HUOQeBTyaOcz3RRQWztwnM09A0AfSTxmutBHasrMSUbE
zumznBNkI2dMPiCdrojrCTmcRJceh8cI/Mx8BIm3+Z0OQiPpZaZxxFxyBQuAf6/z
8TyEgT0B+RNGMZ+771h+8t4ysYt9SPK7IHodyse814F9wxApTc+Ut5KJfnSmN/hu
UvTwssqXX0puFuxPJ61uKkbTUD+y/FqHQmG93cUzC0xYCWDbL/+Rr81MzzMhfBvm
qf/JQl0zI8VXHHz0m1JGXwZRQ6ZsWa5b+hwkYFP8CRkzXvgl/zXKrdvPUuIVo36b
qG8CAwEAAaMnMCUwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMC
MA0GCSqGSIb3DQEBCwUAA4IBAQCrc7O91DuLk17S6iH0786AxH0IEk4fXoHlr0oi
B+tuzO4ccaYxxG0bVwgYIRFqz35YioAoOEdk/xlDbWEn03/ibIqizPkEPwU/JnMY
Jmzph7W9iNkezfUsC1nR3Dw9jVU0TkUz0tnlr57M0jKn3vAzmzMHeil4bnbcmaoC
bnJlJkQ9Uv2OZSfRxc6irVmefC/u97KJPzmGjAlG/KrTVPm/gZtn46szwKKepMqd
7iQJK/xxgHd2kKuOcSVDf2g2ygHbIE7mwofRZLM3VsfaFqIvWBmT1mbC6pzZs30C
m0BmlwDa5ONBGAqjBP9TTm42f4ufzsGF/eFLXZ8lAbpJmLqx
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEcDCCA1igAwIBAgIBAjANBgkqhkiG9w0BAQsFADBfMQ8wDQYDVQQLDAZDNGNv
cmUxDDAKBgNVBAoMA1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwHV2Fs
ZG9yZjELMAkGA1UEBhMCREUxDTALBgNVBAMMBEt5bWEwHhcNMTkwMjEyMTIwNDM4
WhcNMjAwMjEyMTIwNDM4WjBvMQswCQYDVQQGEwJERTEQMA4GA1UECBMHV2FsZG9y
ZjEQMA4GA1UEBxMHV2FsZG9yZjEVMBMGA1UEChMMT3JnYW5pemF0aW9uMRAwDgYD
VQQLEwdPcmdVbml0MRMwEQYDVQQDEwplYy1kZWZhdWx0MIICIjANBgkqhkiG9w0B
AQEFAAOCAg8AMIICCgKCAgEAq4TMGWybHjvhF9RSSZJG8zfp73RJlRhPj3e4EJgr
z0Ai/PmHa8WeK2eVAfsoky9UVE+1t+cn+5trejIzSMf1R8AKpgcvTDkO9RPRLqKm
3u8CjvOrJn0tKSk1Jf9kdSY9xQzd4SOnzSjhbL44MV3zpQ5qdlA2vIvCVNK0SwwS
xqVQbI4FEmbOjtvQKpxnkov2fjviAqcRd+PhR5lNnCGKGtNoVi9lEnXRCdfQsn5C
V67+y8MQkdswBdGrAdSgjTwLvI9kp7eiHAFyRJsLxcFT4VwDFJ2LrEm1hcs/5mDp
UnJWig3g6yVW9ME9oxy7F2etxUsJ4WRFfxixma0hW1AQk1LZduPVmRIHroIDKJA3
79C2gk4b75usaH3gL9s7HUOQeBTyaOcz3RRQWztwnM09A0AfSTxmutBHasrMSUbE
zumznBNkI2dMPiCdrojrCTmcRJceh8cI/Mx8BIm3+Z0OQiPpZaZxxFxyBQuAf6/z
8TyEgT0B+RNGMZ+771h+8t4ysYt9SPK7IHodyse814F9wxApTc+Ut5KJfnSmN/hu
UvTwssqXX0puFuxPJ61uKkbTUD+y/FqHQmG93cUzC0xYCWDbL/+Rr81MzzMhfBvm
qf/JQl0zI8VXHHz0m1JGXwZRQ6ZsWa5b+hwkYFP8CRkzXvgl/zXKrdvPUuIVo36b
qG8CAwEAAaMnMCUwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMC
MA0GCSqGSIb3DQEBCwUAA4IBAQCrc7O91DuLk17S6iH0786AxH0IEk4fXoHlr0oi
B+tuzO4ccaYxxG0bVwgYIRFqz35YioAoOEdk/xlDbWEn03/ibIqizPkEPwU/JnMY
Jmzph7W9iNkezfUsC1nR3Dw9jVU0TkUz0tnlr57M0jKn3vAzmzMHeil4bnbcmaoC
bnJlJkQ9Uv2OZSfRxc6irVmefC/u97KJPzmGjAlG/KrTVPm/gZtn46szwKKepMqd
7iQJK/xxgHd2kKuOcSVDf2g2ygHbIE7mwofRZLM3VsfaFqIvWBmT1mbC6pzZs30C
m0BmlwDa5ONBGAqjBP9TTm42f4ufzsGF/eFLXZ8lAbpJmLqx
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEcDCCA1igAwIBAgIBAjANBgkqhkiG9w0BAQsFADBfMQ8wDQYDVQQLDAZDNGNv
cmUxDDAKBgNVBAoMA1NBUDEQMA4GA1UEBwwHV2FsZG9yZjEQMA4GA1UECAwHV2Fs
ZG9yZjELMAkGA1UEBhMCREUxDTALBgNVBAMMBEt5bWEwHhcNMTkwMjEyMTIwNDM4
WhcNMjAwMjEyMTIwNDM4WjBvMQswCQYDVQQGEwJERTEQMA4GA1UECBMHV2FsZG9y
ZjEQMA4GA1UEBxMHV2FsZG9yZjEVMBMGA1UEChMMT3JnYW5pemF0aW9uMRAwDgYD
VQQLEwdPcmdVbml0MRMwEQYDVQQDEwplYy1kZWZhdWx0MIICIjANBgkqhkiG9w0B
AQEFAAOCAg8AMIICCgKCAgEAq4TMGWybHjvhF9RSSZJG8zfp73RJlRhPj3e4EJgr
z0Ai/PmHa8WeK2eVAfsoky9UVE+1t+cn+5trejIzSMf1R8AKpgcvTDkO9RPRLqKm
3u8CjvOrJn0tKSk1Jf9kdSY9xQzd4SOnzSjhbL44MV3zpQ5qdlA2vIvCVNK0SwwS
xqVQbI4FEmbOjtvQKpxnkov2fjviAqcRd+PhR5lNnCGKGtNoVi9lEnXRCdfQsn5C
V67+y8MQkdswBdGrAdSgjTwLvI9kp7eiHAFyRJsLxcFT4VwDFJ2LrEm1hcs/5mDp
UnJWig3g6yVW9ME9oxy7F2etxUsJ4WRFfxixma0hW1AQk1LZduPVmRIHroIDKJA3
79C2gk4b75usaH3gL9s7HUOQeBTyaOcz3RRQWztwnM09A0AfSTxmutBHasrMSUbE
zumznBNkI2dMPiCdrojrCTmcRJceh8cI/Mx8BIm3+Z0OQiPpZaZxxFxyBQuAf6/z
8TyEgT0B+RNGMZ+771h+8t4ysYt9SPK7IHodyse814F9wxApTc+Ut5KJfnSmN/hu
UvTwssqXX0puFuxPJ61uKkbTUD+y/FqHQmG93cUzC0xYCWDbL/+Rr81MzzMhfBvm
qf/JQl0zI8VXHHz0m1JGXwZRQ6ZsWa5b+hwkYFP8CRkzXvgl/zXKrdvPUuIVo36b
qG8CAwEAAaMnMCUwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMC
MA0GCSqGSIb3DQEBCwUAA4IBAQCrc7O91DuLk17S6iH0786AxH0IEk4fXoHlr0oi
B+tuzO4ccaYxxG0bVwgYIRFqz35YioAoOEdk/xlDbWEn03/ibIqizPkEPwU/JnMY
Jmzph7W9iNkezfUsC1nR3Dw9jVU0TkUz0tnlr57M0jKn3vAzmzMHeil4bnbcmaoC
bnJlJkQ9Uv2OZSfRxc6irVmefC/u97KJPzmGjAlG/KrTVPm/gZtn46szwKKepMqd
7iQJK/xxgHd2kKuOcSVDf2g2ygHbIE7mwofRZLM3VsfaFqIvWBmT1mbC6pzZs30C
m0BmlwDa5ONBGAqjBP9TTm42f4ufzsGF/eFLXZ8lAbpJmLqx
-----END CERTIFICATE-----`
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

func TestCertificateProvider_GetCertificateChain(t *testing.T) {
	t.Run("should get ca certificate", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			certificateChainSecretKey: []byte(pemCertificate),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		certs, err := certificateProvider.GetCertificateChain()

		// then
		require.NoError(t, err)
		require.NotNil(t, certs)
		assert.Equal(t, 3, len(certs))
	})

	t.Run("should return error when failed to read secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(nil, errors.New("error"))

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, err := certificateProvider.GetCertificateChain()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when no data in secret", func(t *testing.T) {
		// given
		secretData := map[string][]byte{}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, err := certificateProvider.GetCertificateChain()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode cert", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			caCertificateSecretKey: []byte("invalid pem"),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, err := certificateProvider.GetCertificateChain()

		// then
		require.Error(t, err)
	})
}

func TestCertificateProvider_GetClientCertificates(t *testing.T) {

	pemPrivateKey := `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	t.Run("should get client certificate and key", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: []byte(pemCertificate),
			clusterKeySecretKey:         []byte(pemPrivateKey),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		key, cert, err := certificateProvider.GetClientCredentials()

		// then
		require.NoError(t, err)
		require.NotNil(t, cert)
		require.NotNil(t, key)
	})

	t.Run("should return error when failed to read secret data", func(t *testing.T) {
		// given
		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(nil, errors.New("error"))

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, _, err := certificateProvider.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when no cert in secret", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterKeySecretKey: []byte(pemPrivateKey),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, _, err := certificateProvider.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when no key in secret", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: []byte(pemCertificate),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, _, err := certificateProvider.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode cert", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: []byte("invalid pem"),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, _, err := certificateProvider.GetClientCredentials()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to decode cert", func(t *testing.T) {
		// given
		secretData := map[string][]byte{
			clusterCertificateSecretKey: []byte(pemCertificate),
			clusterKeySecretKey:         []byte("invalid pem"),
		}

		secretRepository := &mocks.Repository{}
		secretRepository.On("Get", clusterCertSecretNamespaceName).Return(secretData, nil)

		certificateProvider := NewCertificateProvider(clusterCertSecretNamespaceName, caCertSecretNamespaceName, secretRepository)

		// when
		_, _, err := certificateProvider.GetClientCredentials()

		// then
		require.Error(t, err)
	})
}
