package connectorservice

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates/mocks"

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
)

func TestMutualTLSClientProvider_CreateClient(t *testing.T) {

	t.Run("should create Mutual TLS Connector Client", func(t *testing.T) {
		// given
		csrProvider := &mocks.CSRProvider{}

		key := loadPrivateKey(t, []byte(clientKey))
		cert := loadCertificates(t, []byte(clientCertificate))

		clientProvider := NewMutualTLSClientProvider(csrProvider)

		// when
		client := clientProvider.CreateClient(key, cert)

		// then
		require.NotNil(t, client)
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

func loadCertificates(t *testing.T, certificate []byte) *x509.Certificate {
	pemBlock, _ := pem.Decode(certificate)
	require.NotNil(t, pemBlock)

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	require.NoError(t, err)

	return cert
}
