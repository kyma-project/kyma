package testkit

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	rsaKeySize = 2048
)

// Create Key generates rsa.PrivateKey
func CreateKey(t *testing.T) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	require.NoError(t, err)

	return key
}

// CreateCsr creates CSR request
func CreateCsr(t *testing.T, certInfo CertInfo, keys *rsa.PrivateKey) []byte {
	subjectInfo := extractSubject(certInfo.Subject)

	subject := pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}

	var csrTemplate = x509.CertificateRequest{
		Subject: subject,
	}

	// step: generate the csr request
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, keys)
	require.NoError(t, err)

	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	return csr
}

// EncodedCertChainToPemBytes decodes certificates chain and return pemBlock's bytes for client cert and ca cert
func EncodedCertChainToPemBytes(t *testing.T, encodedChain string) []byte {
	crtBytes := decodeBase64Cert(encodedChain, t)

	clientCrtPem, rest := pem.Decode(crtBytes)
	require.NotNil(t, clientCrtPem)
	require.NotEmpty(t, rest)

	caCrtPem, _ := pem.Decode(rest)
	require.NotNil(t, caCrtPem)

	certChainBytes := append(clientCrtPem.Bytes, caCrtPem.Bytes...)

	return certChainBytes
}

// EncodedCertToPemBytes decodes certificate and return pemBlock's bytes for it
func EncodedCertToPemBytes(t *testing.T, encodedCert string) []byte {
	crtBytes := decodeBase64Cert(encodedCert, t)

	certificate, _ := pem.Decode(crtBytes)
	require.NotNil(t, certificate)

	return certificate.Bytes
}

// DecodeAndParseCerts decodes base64 encoded certificates chain and parses it
func DecodeAndParseCerts(t *testing.T, crtResponse *CrtResponse) DecodedCrtResponse {
	certChainBytes := EncodedCertChainToPemBytes(t, crtResponse.CRTChain)
	certificateChain, err := x509.ParseCertificates(certChainBytes)
	require.NoError(t, err)

	clientCertBytes := EncodedCertToPemBytes(t, crtResponse.ClientCRT)
	clientCertificate, err := x509.ParseCertificate(clientCertBytes)
	require.NoError(t, err)

	caCertificateBytes := EncodedCertToPemBytes(t, crtResponse.CaCRT)
	caCertificate, err := x509.ParseCertificate(caCertificateBytes)
	require.NoError(t, err)

	return DecodedCrtResponse{
		CRTChain:  certificateChain,
		ClientCRT: clientCertificate,
		CaCRT:     caCertificate,
	}
}

// CheckIfSubjectEquals verifies that specified subject is equal to this in certificate
func CheckIfSubjectEquals(t *testing.T, expectedSubject string, certificate *x509.Certificate) {
	subjectInfo := extractSubject(expectedSubject)
	actualSubject := certificate.Subject

	require.Equal(t, subjectInfo["CN"], actualSubject.CommonName)
	require.Equal(t, []string{subjectInfo["C"]}, actualSubject.Country)
	require.Equal(t, []string{subjectInfo["O"]}, actualSubject.Organization)
	require.Equal(t, []string{subjectInfo["OU"]}, actualSubject.OrganizationalUnit)
	require.Equal(t, []string{subjectInfo["L"]}, actualSubject.Locality)
	require.Equal(t, []string{subjectInfo["ST"]}, actualSubject.Province)
}

// CheckIfCertIsSigned verifies that client certificate is signed by server certificate
func CheckIfCertIsSigned(t *testing.T, certificates []*x509.Certificate) {
	clientCrt := certificates[0]
	serverCrt := certificates[1]

	err := clientCrt.CheckSignatureFrom(serverCrt)

	require.NoError(t, err)
}

func EncodeBase64(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func decodeBase64Cert(certificate string, t *testing.T) []byte {
	crtBytes, err := base64.StdEncoding.DecodeString(certificate)
	require.NoError(t, err)
	return crtBytes
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
