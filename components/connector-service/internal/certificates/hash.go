package certificates

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

func CalculateHash(cert string) (string, error) {
	unescapedCert, err := url.PathUnescape(cert)
	if err != nil {
		return "", err
	}
	input := []byte(unescapedCert)
	sha := sha256.Sum256(input)

	return hex.EncodeToString(sha[:]), nil
}
