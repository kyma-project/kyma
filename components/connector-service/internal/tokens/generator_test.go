package tokens

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomString(t *testing.T) {

	t.Run("should generate random string", func(t *testing.T) {
		cases := []struct {
			length int
		}{
			{10}, {20}, {5},
		}

		for _, elem := range cases {
			randStr, apperr := GenerateRandomString(elem.length)
			require.NoError(t, apperr)

			decoded, err := base64.URLEncoding.DecodeString(randStr)
			require.NoError(t, err)

			assert.Equal(t, elem.length, len(decoded))
		}
	})
}
