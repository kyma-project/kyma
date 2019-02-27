package certificates

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateHash(t *testing.T) {

	t.Run("should calculate correct hash", func(t *testing.T) {
		// given
		hash := sha256.New()
		hash.Write([]byte(cert))

		expectedHash := hex.EncodeToString(hash.Sum(nil))

		// when
		calculatedHash := CalculateHash(cert)

		// then
		assert.Equal(t, calculatedHash, expectedHash)
	})
}