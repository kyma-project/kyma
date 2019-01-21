package certificates

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
)

const (
	rsaKeySize              = 2048
	certificateValidityDays = 365
)

type Generator func(pkix.Name) (*KeyCertPair, apperrors.AppError)

func GenerateKeyAndCertificate(subject pkix.Name) (*KeyCertPair, apperrors.AppError) {
	key, apperr := generateKey()
	if apperr != nil {
		return nil, apperr
	}

	certBytes, err := generateCertificate(subject, key)
	if err != nil {
		return nil, apperrors.Internal("Failed to generate certificate, %s", err)
	}

	encodedCrt, apperr := encodePemBlock(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if apperr != nil {
		return nil, apperr
	}

	encodedKey, apperr := encodePemBlock(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if apperr != nil {
		return nil, apperr
	}

	return &KeyCertPair{
		Certificate: encodedCrt,
		PrivateKey:  encodedKey,
	}, nil
}

func generateKey() (*rsa.PrivateKey, apperrors.AppError) {
	key, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return nil, apperrors.Internal("Failed to generate private key, %s", err.Error())
	}

	return key, nil
}

func generateCertificate(subject pkix.Name, key *rsa.PrivateKey) ([]byte, error) {
	template := x509.Certificate{
		SignatureAlgorithm: x509.SHA256WithRSA,

		SerialNumber: big.NewInt(2),
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(certificateValidityDays * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	return x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
}

func encodePemBlock(block *pem.Block) ([]byte, apperrors.AppError) {
	buffer := &bytes.Buffer{}
	err := pem.Encode(buffer, block)
	if err != nil {
		return nil, apperrors.Internal("Failed to encode private key, %s", err)
	}

	return buffer.Bytes(), nil
}
