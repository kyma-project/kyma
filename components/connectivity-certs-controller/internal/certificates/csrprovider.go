package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/secrets"

	"github.com/pkg/errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	rsaKeySize          = 4096
	clusterKeySecretKey = "key"
)

type CSRProvider interface {
	CreateCSR(plainSubject string) (string, error)
}

type csrProvider struct {
	clusterCertSecretName string
	caCRTSecretName       string
	secretRepository      secrets.Repository
}

func NewCSRProvider(clusterCertSecret, caCRTSecret string, secretRepository secrets.Repository) CSRProvider {
	return &csrProvider{
		clusterCertSecretName: clusterCertSecret,
		caCRTSecretName:       caCRTSecret,
		secretRepository:      secretRepository,
	}
}

func (cp *csrProvider) CreateCSR(plainSubject string) (string, error) {
	clusterPrivateKey, err := cp.provideClusterPrivateKey()
	if err != nil {
		return "", err
	}

	csr, err := createCSR(plainSubject, clusterPrivateKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(csr), nil
}

func createCSR(plainSubject string, key *rsa.PrivateKey) ([]byte, error) {
	subject := parseSubject(plainSubject)

	csrTemplate := x509.CertificateRequest{
		Subject: subject,
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create cluster CSR")
	}

	pemEncodedCSR := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})

	return pemEncodedCSR, nil
}

func parseSubject(plainSubject string) pkix.Name {
	subjectInfo := extractSubject(plainSubject)

	return pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}
}

func extractSubject(plainSubject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(plainSubject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}

func (cp *csrProvider) provideClusterPrivateKey() (*rsa.PrivateKey, error) {
	secret, err := cp.secretRepository.Get(cp.clusterCertSecretName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return cp.createClusterKeySecret()
		}
		return nil, errors.Wrapf(err, fmt.Sprintf("Failed to read cluster %s secret", cp.clusterCertSecretName))
	}

	block, _ := pem.Decode(secret[clusterKeySecretKey])
	if block == nil {
		return cp.createClusterKeySecret()
	}

	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return privateKey, nil
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error while parsing private key")
	}

	return privateKey.(*rsa.PrivateKey), nil
}

func (cp *csrProvider) createClusterKeySecret() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, err
	}

	secretData := map[string][]byte{
		clusterKeySecretKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
	}

	err = cp.secretRepository.UpsertWithReplace(cp.clusterCertSecretName, secretData)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to override cluster key secret")
	}

	return key, nil
}
