package connectorservice

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	clientKey = `-----BEGIN RSA PRIVATE KEY-----
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

	clientCertificate = `-----BEGIN CERTIFICATE-----
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

	application   = "test-app"
	tenant        = "test-tenant"
	group         = "test-group"
	eventsURL     = "https://gateway.events/"
	metadataURL   = "https://gateway.metadata/"
	revocationURL = "https://gateway.revocation/"
	renewalURL    = "https://gateway.renewal/"
)

func TestMutualTLSConnectorClient_GetManagementInfo(t *testing.T) {

	managementInfoEndpoint := "/v1/application/management/info"

	t.Run("should get management info", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(managementInfoEndpoint, func(w http.ResponseWriter, r *http.Request) {
			managementInfo := ManagementInfo{
				ClientIdentity: ClientIdentity{
					Application: application,
					Tenant:      tenant,
					Group:       group,
				},
				ManagementURLs: ManagementURLs{
					EventsURL:     eventsURL,
					MetadataURL:   metadataURL,
					RevocationURL: revocationURL,
					RenewalURL:    renewalURL,
				},
			}

			respond(t, w, http.StatusOK, managementInfo)
		})

		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		managementInfoResponse, err := mutualTLSClient.GetManagementInfo(server.URL + managementInfoEndpoint)

		// then
		require.NoError(t, err)

		assert.Equal(t, application, managementInfoResponse.ClientIdentity.Application)
	})

	t.Run("should return error when request failed", func(t *testing.T) {
		// given
		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		_, err := mutualTLSClient.GetManagementInfo("https://invalid.url.kyma.cx")

		// then
		require.Error(t, err)
	})

	t.Run("should return error when server responded with error", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(managementInfoEndpoint, errorHandler(t))

		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		_, err := mutualTLSClient.GetManagementInfo(server.URL + managementInfoEndpoint)

		// then
		require.Error(t, err)
	})

}

func TestMutualTLSConnectorClient_RenewCertificate(t *testing.T) {

	encodedCSR := "encodedCSR"
	renewalEndpoint := "/v1/application/certificates/renewals"

	t.Run("should renew certificate", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(renewalEndpoint, func(w http.ResponseWriter, r *http.Request) {
			var certRequest CertificateRequest
			err := readResponseBody(r.Body, &certRequest)
			require.NoError(t, err)
			assert.Equal(t, encodedCSR, certRequest.CSR)

			crtResponse := CertificatesResponse{
				CRTChain:  crtChainBase64,
				ClientCRT: clientCRTBase64,
				CaCRT:     caCRTBase64,
			}

			respond(t, w, http.StatusCreated, crtResponse)
		})

		renewalURL := fmt.Sprintf("%s%s", server.URL, renewalEndpoint)

		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		certificates, err := mutualTLSClient.RenewCertificate(renewalURL, encodedCSR)
		require.NoError(t, err)

		// then
		assert.Equal(t, clientCRT, certificates.ClientCRT)
		assert.Equal(t, caCRT, certificates.CaCRT)
		assert.Equal(t, crtChain, certificates.CRTChain)
	})

	t.Run("should return error when request failed", func(t *testing.T) {
		// given
		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		_, err := mutualTLSClient.RenewCertificate("https://invalid.url.kyma.cx", encodedCSR)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when server responded with error", func(t *testing.T) {
		// given
		server, router := createTestServer()
		defer server.Close()

		router.HandleFunc(renewalEndpoint, errorHandler(t))

		renewalURL := fmt.Sprintf("%s%s", server.URL, renewalEndpoint)

		mutualTLSClient := NewMutualTLSConnectorClient(loadPrivateKey(t, []byte(clientKey)), loadCertificates(t, []byte(clientCertificate)))

		// when
		_, err := mutualTLSClient.RenewCertificate(renewalURL, encodedCSR)

		// then
		require.Error(t, err)
	})
}

func loadPrivateKey(t *testing.T, key []byte) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(key))
	require.NotNil(t, block)

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err)

	return privateKey.(*rsa.PrivateKey)
}

func loadCertificates(t *testing.T, certificate []byte) []*x509.Certificate {
	pemBlock, rest := pem.Decode(certificate)
	require.NotNil(t, pemBlock)

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	require.NoError(t, err)

	var certificates []*x509.Certificate
	certificates = append(certificates, cert)

	pemBlock2, _ := pem.Decode(rest)
	if pemBlock2 != nil {
		cert2, err := x509.ParseCertificate(pemBlock2.Bytes)
		require.NoError(t, err)

		certificates = append(certificates, cert2)
	}

	return certificates
}
