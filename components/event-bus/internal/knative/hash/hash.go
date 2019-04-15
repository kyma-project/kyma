package hash

import (
	"crypto/sha1"
	"encoding/base32"
)

const (
	encoder = "abcdefghijklmnopqrstuvwxyz234567"
)

var (
	base32Encoding = base32.NewEncoding(encoder)
)

func ComputeHash(input *string) string {
	h := sha1.New()
	h.Write([]byte(*input))
	b := h.Sum(nil)
	return base32Encoding.EncodeToString(b)
}
