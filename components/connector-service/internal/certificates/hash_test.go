package certificates

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCalculateHash(t *testing.T) {

	t.Run("should calculate correct hash", func(t *testing.T) {
		// given
		testCert := "testCert%0ACorrectEscape"
		expectedHash := "afc8b317a038c53d50fb5799cc1cd9270f0009381561210fb9c06c7fc73c93e6"

		// when
		calculatedHash, err := CalculateHash(testCert)
		require.NoError(t, err)

		// then
		assert.Equal(t, expectedHash, calculatedHash)
	})

	t.Run("should return error, when unable to unescape cert", func(t *testing.T) {
		// given
		testCert := "testCert%WrongEscape%"

		// when
		hash, err := CalculateHash(testCert)

		// then
		assert.Equal(t, "", hash)
		assert.Error(t, err)
	})
}
