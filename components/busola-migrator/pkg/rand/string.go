package rand

import (
	"crypto/rand"
	"encoding/hex"
)

func Hex(n int) (string, error) {
	b, err := randBytes(n)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randBytes(n int) ([]byte, error) {
	var randomBytes = make([]byte, n)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	return randomBytes, nil
}
