package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/application-connectivity-certs-setup-job/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	caSecretName      = "application-connector-ca-certs"
	caSecretNamespace = "istio-system"

	connectorSecretName      = "connector-service-app-ca"
	connectorSecretNamespace = "kyma-integration"
)

const (
	privateKeyPem = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA6w9QEKbP8EhFE6zAPIVlFywZEKqSIABUDzoPZHLSErIru/Np
iI8+jpv/CxeL7k03+bP8BnUfJ31hoTL3QDFIvba5+RiixDdjHAweT2u0+5AP+ePw
x5JIASdpYCKDsYeg8o7B0EOYTPwT0gmf8aV8rOaND1ZvJ66H7RhquiI/6V34aWeD
7IMYpgQsrtKSq3wrN6ICZDxsn9bkpny81h8W/At+PlBTGz5cufOU6amnsuOMvGwe
oy99G+jaoPAM5GOA/d2AVBiZ8OU59ZP5Ur8Erl8Fvp+ynPHEaCWGCVV5RGRPsvQa
qnvYakgJXMvBn6MwUf3AoqNToS87Kq1dCmeajQIDAQABAoIBAQCIlnNN2cDGvRf2
oNFr2Y+ucV93QcZ7dfVii7haBCZx2rpzErRmN+Z/88G17k7PgGtgW+e80N3zknXi
t7zYvkqogr96MYiTQCQFLj2GpO2bqFDAQmWtciEJGp+uzx97T3aEu9N/c2fShD/4
MsOQJTtXNPkOyoj4pAA0E5Yg5roAng7edJwO9fujSbhmJtvnVLz2ZC/06TuC0tDU
cIGsJYqYYGG0lTE41iecYJs1mDwxIofWsl3GHYEU3dMILaWzfvWRAl+msx8TzyOw
4WXgOpcwV7uOtPAHN/OWpjocKbzW6ouqraL8Quft4ioYm9qeRH9mdB13d1H1Sqnc
kMK6Wrm5AoGBAP/O9vHFlbUxhKWvwqcfbb38zaVXbt0TPpnqzthpKxXLFqoPYKR0
fZQs3Gt36Z/pqNioQfzqwvthd9NgzHbNvPlPbvwqHS3C0DsZ22O0VvLPmmbybPtZ
aIExZujZ42iqXpBpv9yr9L7UGyvvAdb5D899+yXqEKb79qnU063RpUKvAoGBAOs8
XvEFQvW2C5UkGANv3zvhWPdHp99bPL/ePI2OIF35NUulbnrxZu39sCo/6hmog3kP
1YH+lT/fsMwNuvlepGztO29+CHOz6KiLGXYwCTwpLZjTREUJba6eiefhnYnGKlpn
AIy2pzs/LJUveRepojJNnthvUjKXLjiULPcri/WDAoGBAKqynK5wvpmOVYmKY0XJ
/x0MGN4AHgZ/1QI4YZafdxSv1Ivefwq+gR3jYaKE/eyrqvQIMyBmN34vaBoxOb79
QuDKVLEIGTh0Cyek9XTu3iZgyhNwKbD/1HCBWr5+xvUM2tVa+6BxTnwYZZlHf97H
i/lVg8WlDz+eWtaxIh+XCcQZAoGAe+XCQ8P3rp8Bnr3x/+1ucIWSbDvLiXLunkgZ
MJ2JIrXdgkhR1mNLSVJy9O3RCU6eYKccV2mVhpz066TXs/xLMiwJQAHrxbUed5c8
A+ntE0jFAVdU/9+la3GJRR6p8ST0rcTOn06c6jGt862bZAEusrv7TBfl/UtvRtGU
lWLURq0CgYEA/QhiaD6alc01yk2bazVt5Ofl3Eqw7yATO21EPymM51LaLMyAw6o0
nXLJnKSbO+eB020eCGMDgcR5FjE46NWcOMLdCINZ2iFkhIJo+M/lEAAXvqVLRcQr
sqXzf1JPPHnmozIitzXi9prXUE3xQ66i95l2GQq2OHHIMyHljtYk9bQ=
-----END RSA PRIVATE KEY-----`

	certificatePem = `-----BEGIN CERTIFICATE-----
MIICIzCCAYwCCQDDkk/CKHDcZjANBgkqhkiG9w0BAQUFADASMRAwDgYDVQQKEwdB
Y21lIENvMCAXDTE5MDMyOTEzMjU1M1oYDzIxMTkwMzA1MTMyNTUzWjAUMRIwEAYD
VQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDr
D1AQps/wSEUTrMA8hWUXLBkQqpIgAFQPOg9kctISsiu782mIjz6Om/8LF4vuTTf5
s/wGdR8nfWGhMvdAMUi9trn5GKLEN2McDB5Pa7T7kA/54/DHkkgBJ2lgIoOxh6Dy
jsHQQ5hM/BPSCZ/xpXys5o0PVm8nroftGGq6Ij/pXfhpZ4PsgximBCyu0pKrfCs3
ogJkPGyf1uSmfLzWHxb8C34+UFMbPly585Tpqaey44y8bB6jL30b6Nqg8AzkY4D9
3YBUGJnw5Tn1k/lSvwSuXwW+n7Kc8cRoJYYJVXlEZE+y9Bqqe9hqSAlcy8GfozBR
/cCio1OhLzsqrV0KZ5qNAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAi9t6j7ahK9vZ
VsfqyMGcgeIrI2mzI8oDAHb0xkrKiQpOAGoq9ejBujwDI3L2g2MToHhB0aataCmC
oiCU2Sf1LDG70bnyd0eLKshNEFjHEsVHJkzPwxeOFsM7xuKCZQ4uvnFBZyyQmuyY
QbIjsJhuMRQuka2NB6eGq4qFaHHbkzc=
-----END CERTIFICATE-----`

	notMatchingCertificate = `-----BEGIN CERTIFICATE-----
MIICmjCCAYICCQD5MM5a/ZNUQDANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDDARL
eW1hMB4XDTE5MDcwMzEwMTYxNloXDTIwMDcwMjEwMTYxNlowDzENMAsGA1UEAwwE
S3ltYTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAOSOavHRErvHIBmX
LUcva2rTu/SaqPERrz8S8IE2pQX6mJRwDZSJuGFUm803SBeuo/DPDjnqWYM2Grf0
vK1ftjs8dQ61hYVbkeAefcCFVBGtHeTtpkm/YqL6pa8iRpX2ZBG5A92GVDBlfhS8
yzeAgCeTnxodPVKRW7uoSsei8/+iOI0RAR07P9lPNvbiNW1q0hox0TF0Bo3VwrHP
CbhHNbFX4YBG9xCGLwyt4T1yjP4dcUkeouMRB2tsg40c85zjFXUEAGs46LQTasfc
GpH/z5Wq925cR7/0ykW88AEnuY2wt65XXFmCLcJjJNVyyfVifqXawiuYg5kAFb8U
wkvvJNsCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAdxZsRi/1qHrrA/+pmpZgwz9G
BmECYujN3Hos6hz768vLAdK2BfkEplH/VY9qg/PLbuhGIL76x8AQEAlA2859Zxrw
3q7QJE0wZxU95Kh43E3VY1ArF+/GHq/zu/Am44zgQsM4yyeZal9qyxcgLeAxG97+
kas7i+NCwSlU4kYjBy5O/5upa1do+MLqLIKezLC1wAVlw0pZmCuWXrGgP4gwx++9
ob8w/gmij/nh5zApgrrhUXBccl4TntGsR664jMxU0I7eJySGfshjXCbh7oZEOVHm
ibwixBXWcQIr7xX801h/MkAtVLH4e5vy3YLLojYMp3ak8VQubcLGNKlHUP86Bg==
-----END CERTIFICATE-----`

	invalidCertificate = `-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`

	invalidKey = `-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`
)

var (
	caSecretNamespacedName        = types.NamespacedName{Name: caSecretName, Namespace: caSecretNamespace}
	connectorSecretNamespacedName = types.NamespacedName{Name: connectorSecretName, Namespace: connectorSecretNamespace}
)

func TestCertSetupHandler_SetupApplicationConnectorCertificate(t *testing.T) {

	t.Run("should not modify existing certificates if new ones not provided", func(t *testing.T) {
		// given
		existingCaSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      caSecretNamespacedName.Name,
				Namespace: caSecretNamespacedName.Namespace,
			},
			Data: map[string][]byte{
				"cacert": []byte("cacert"),
			},
		}
		existingConnectorSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      connectorSecretNamespacedName.Name,
				Namespace: connectorSecretNamespacedName.Namespace,
			},
			Data: map[string][]byte{
				"ca.crt": []byte("cacert"),
				"ca.key": []byte("key"),
			},
		}

		fakeClientSet := fake.NewSimpleClientset(existingCaSecret, existingConnectorSecret).CoreV1()

		secretRepository := NewSecretRepository(func(namespace string) Manager {
			return fakeClientSet.Secrets(namespace)
		})

		options := &options{
			connectorCertificateSecret: connectorSecretNamespacedName,
			caCertificateSecret:        caSecretNamespacedName,
			caCertificate:              "",
			caKey:                      "",
		}

		certSetupHandler := NewCertificateSetupHandler(options, secretRepository)

		// when
		err := certSetupHandler.SetupApplicationConnectorCertificate()

		// then
		require.NoError(t, err)

		caSecret, err := secretRepository.Get(caSecretNamespacedName)
		require.NoError(t, err)
		assert.EqualValues(t, existingCaSecret.Data, caSecret)

		connectorSecret, err := secretRepository.Get(connectorSecretNamespacedName)
		require.NoError(t, err)
		assert.EqualValues(t, existingConnectorSecret.Data, connectorSecret)
	})

	t.Run("should update secrets with provided certificates", func(t *testing.T) {
		// given
		secretRepository := fakeRepositoryWithEmptySecrets()

		options := &options{
			connectorCertificateSecret: connectorSecretNamespacedName,
			caCertificateSecret:        caSecretNamespacedName,
			caCertificate:              base64Encode(certificatePem),
			caKey:                      base64Encode(privateKeyPem),
		}

		certSetupHandler := NewCertificateSetupHandler(options, secretRepository)

		// when
		err := certSetupHandler.SetupApplicationConnectorCertificate()

		// then
		require.NoError(t, err)

		caSecret, err := secretRepository.Get(caSecretNamespacedName)
		require.NoError(t, err)
		assert.Equal(t, []byte(certificatePem), caSecret["cacert"])

		connectorSecret, err := secretRepository.Get(connectorSecretNamespacedName)
		require.NoError(t, err)
		assert.Equal(t, []byte(certificatePem), connectorSecret["ca.crt"])
		assert.Equal(t, []byte(privateKeyPem), connectorSecret["ca.key"])
	})

	t.Run("should return error when failed to upsert ca secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.SecretRepository{}
		secretRepository.On("Get", connectorSecretNamespacedName).Return(emptySecret(connectorSecretNamespacedName), nil)
		secretRepository.On("Get", caSecretNamespacedName).Return(emptySecret(caSecretNamespacedName), nil)

		secretRepository.On("Upsert", caSecretNamespacedName, mock.Anything).
			Return(errors.New("error"))

		options := &options{
			connectorCertificateSecret: connectorSecretNamespacedName,
			caCertificateSecret:        caSecretNamespacedName,
			caCertificate:              base64Encode(certificatePem),
			caKey:                      base64Encode(privateKeyPem),
		}

		certSetupHandler := NewCertificateSetupHandler(options, secretRepository)

		// when
		err := certSetupHandler.SetupApplicationConnectorCertificate()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to upsert connector secret", func(t *testing.T) {
		// given
		secretRepository := &mocks.SecretRepository{}
		secretRepository.On("Get", connectorSecretNamespacedName).Return(emptySecret(connectorSecretNamespacedName), nil)
		secretRepository.On("Get", caSecretNamespacedName).Return(emptySecret(caSecretNamespacedName), nil)

		secretRepository.On("Upsert", caSecretNamespacedName, mock.Anything).
			Return(nil)
		secretRepository.On("Upsert", connectorSecretNamespacedName, mock.Anything).
			Return(errors.New("error"))

		options := &options{
			connectorCertificateSecret: connectorSecretNamespacedName,
			caCertificateSecret:        caSecretNamespacedName,
			caCertificate:              base64Encode(certificatePem),
			caKey:                      base64Encode(privateKeyPem),
		}

		certSetupHandler := NewCertificateSetupHandler(options, secretRepository)

		// when
		err := certSetupHandler.SetupApplicationConnectorCertificate()

		// then
		require.Error(t, err)
	})

}

func TestCertSetupHandler_SetupApplicationConnectorCertificate_GeneratingCertificates(t *testing.T) {
	validityTime := time.Minute * 60

	for _, test := range []struct {
		description string
		caCert      string
		caKey       string
	}{
		{
			description: "certificate and key not provided",
			caCert:      "",
			caKey:       "",
		},
		{
			description: "invalid base64 certificate",
			caCert:      "invalid base 64",
			caKey:       base64Encode(privateKeyPem),
		},
		{
			description: "invalid base64 certificate",
			caCert:      base64Encode(certificatePem),
			caKey:       "invalid base 64",
		},
		{
			description: "certificate not provided",
			caCert:      "",
			caKey:       base64Encode(privateKeyPem),
		},
		{
			description: "key not provided",
			caCert:      base64Encode(certificatePem),
			caKey:       "",
		},
		{
			description: "certificate is invalid pem",
			caCert:      base64Encode("invalid"),
			caKey:       base64Encode(privateKeyPem),
		},
		{
			description: "key is invalid pem",
			caCert:      base64Encode(certificatePem),
			caKey:       base64Encode("invalid"),
		},
		{
			description: "certificate is invalid",
			caCert:      base64Encode(invalidCertificate),
			caKey:       base64Encode(privateKeyPem),
		},
		{
			description: "key is invalid",
			caCert:      base64Encode(certificatePem),
			caKey:       base64Encode(invalidKey),
		},
		{
			description: "key and certificate does not match",
			caCert:      base64Encode(notMatchingCertificate),
			caKey:       base64Encode(privateKeyPem),
		},
	} {
		t.Run("should generate key and certificate when "+test.description, func(t *testing.T) {
			secretRepository := fakeRepositoryWithEmptySecrets()

			options := &options{
				connectorCertificateSecret: connectorSecretNamespacedName,
				caCertificateSecret:        caSecretNamespacedName,
				caCertificate:              test.caCert,
				caKey:                      test.caKey,
				generatedValidityTime:      validityTime,
			}

			certSetupHandler := NewCertificateSetupHandler(options, secretRepository)

			// when
			err := certSetupHandler.SetupApplicationConnectorCertificate()

			// then
			require.NoError(t, err)
			assertCertificateAndKeyGenerated(t, validityTime, secretRepository)
		})
	}
}

func assertCertificateAndKeyGenerated(t *testing.T, validityTime time.Duration, secretRepository SecretRepository) {
	caSecret, err := secretRepository.Get(caSecretNamespacedName)
	require.NoError(t, err)
	caCert := assertCertificate(t, caSecret["cacert"])

	connectorSecret, err := secretRepository.Get(connectorSecretNamespacedName)
	require.NoError(t, err)
	connectorCaCert := assertCertificate(t, connectorSecret["ca.crt"])
	assertKey(t, connectorSecret["ca.key"])

	assert.Equal(t, caCert, connectorCaCert)
	assert.True(t, caCert.NotAfter.Unix() <= time.Now().Add(validityTime).Unix())
}

func assertCertificate(t *testing.T, cert []byte) *x509.Certificate {
	block, _ := pem.Decode(cert)
	require.NotNil(t, block)

	certificate, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	assert.NotNil(t, certificate)

	return certificate
}

func assertKey(t *testing.T, key []byte) *rsa.PrivateKey {
	block, _ := pem.Decode(key)
	require.NotNil(t, block)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err)

	assert.NotNil(t, privateKey)

	return privateKey
}

func fakeRepositoryWithEmptySecrets() SecretRepository {
	caSecret := emptySecret(caSecretNamespacedName)
	connectorSecret := emptySecret(connectorSecretNamespacedName)

	fakeClientSet := fake.NewSimpleClientset(caSecret, connectorSecret).CoreV1()

	return NewSecretRepository(func(namespace string) Manager {
		return fakeClientSet.Secrets(namespace)
	})
}

func emptySecret(name types.NamespacedName) *v1.Secret {
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Data: map[string][]byte{},
	}
}

func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
