package testkit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/pkg/errors"
)

const (
	rsaKeySize = 2048
	csrType    = "CERTIFICATE REQUEST"
)

func CreateKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func CreateCSR(subjectRaw string, key *rsa.PrivateKey) ([]byte, error) {
	subjectInfo := extractSubject(subjectRaw)

	subject := pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}

	csrTemplate := &x509.CertificateRequest{
		Subject: subject,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, key)
	if err != nil {
		return nil, err
	}

	csr := pem.EncodeToMemory(&pem.Block{
		Type:  csrType,
		Bytes: csrBytes,
	})

	return csr, nil
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

func extractSubject(subject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(subject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}
