package certificates

import (
	"crypto/sha256"
	"fmt"
)

func CalculateHash(cert string) string {
	input := []byte(cert)
	sha := sha256.Sum256(input)

	hexified := ""
	for _, data := range sha {
		hexified += fmt.Sprintf("%02x", data)
	}
	return hexified
}