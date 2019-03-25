package certificates

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

func FingerprintSHA256(rawPem []byte) (string, error) {
	block, _ := pem.Decode(rawPem)
	if block == nil {
		return "", apperrors.Internal("Faield to decode pem block")
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", apperrors.Internal("Failed to parse certificate: %s", err.Error())
	}

	sha := sha256.Sum256(certificate.Raw)

	return hex.EncodeToString(sha[:]), nil
}
