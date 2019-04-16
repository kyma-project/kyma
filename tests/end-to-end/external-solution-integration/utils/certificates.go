package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	rsaKeySize = 2048
)

func createKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	require.NoError(t, err)

	return key
}

func createCSR(t *testing.T, subject Subject, key *rsa.PrivateKey) []byte {
	sub := pkix.Name{
		CommonName:         subject.CommonName,
		Country:            []string{subject.Country},
		Organization:       []string{subject.Organization},
		OrganizationalUnit: []string{subject.OrganizationalUnit},
		Locality:           []string{subject.Locality},
		Province:           []string{subject.Province},
	}

	csrTemplate := &x509.CertificateRequest{
		Subject: sub,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, key)
	require.NoError(t, err)

	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	return csr
}

func encodedCertChainToPemBytes(encodedChain string) ([]byte, error) {
	crtBytes, err := decodeBase64Cert(encodedChain)
	if err != nil {
		return nil, err
	}

	clientCrtPem, rest := pem.Decode(crtBytes)
	if clientCrtPem == nil {
		return nil, errors.New("error while decoding client certificate pem block")
	}

	caCrtPem, _ := pem.Decode(rest)
	if clientCrtPem == nil {
		return nil, errors.New("error while decoding ca certificate pem block")
	}

	certChainBytes := append(clientCrtPem.Bytes, caCrtPem.Bytes...)

	return certChainBytes, nil
}

func decodeBase64Cert(certificate string) ([]byte, error) {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	if err != nil {
		return nil, err
	}
	return crtBytes, nil
}
