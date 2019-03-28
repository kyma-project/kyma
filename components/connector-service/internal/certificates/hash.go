package certificates

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
)

// FingerprintSHA256 decodes pem block and returns fingerprint generated using SHA 256 algorithm
func FingerprintSHA256(rawPem []byte) (string, apperrors.AppError) {
	block, _ := pem.Decode(rawPem)
	if block == nil {
		return "", apperrors.Internal("Failed to decode pem block.")
	}

	sha := sha256.Sum256(block.Bytes)

	return hex.EncodeToString(sha[:]), nil
}
