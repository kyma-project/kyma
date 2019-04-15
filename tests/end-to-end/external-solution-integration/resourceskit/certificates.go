package resourceskit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	rsaKeySize = 2048
)

func CreateKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	require.NoError(t, err)

	return key
}

func CreateCSR(t *testing.T, subject Subject, key *rsa.PrivateKey) []byte {
	sub := pkix.Name{
		CommonName: subject.CommonName,
		Country: []string{subject.Country},
		Organization: []string{subject.Organization},
		OrganizationalUnit:[]string{subject.OrganizationalUnit},
		Locality:[]string{subject.Locality},
		Province:[]string{subject.Province},
	}

	csrTemplate := &x509.CertificateRequest{
		Subject:sub,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, key)
	require.NoError(t, err)

	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	return csr
}