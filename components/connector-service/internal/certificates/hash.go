package certificates

import (
	"crypto/sha256"
	"encoding/hex"
)

func CalculateHash(cert string) string {
	input := []byte(cert)
	sha := sha256.Sum256(input)

	return hex.EncodeToString(sha[:])
}