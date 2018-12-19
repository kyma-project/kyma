package tokens

import (
	"encoding/base64"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	tokenLength = 10
	reName      = "appName"
)

func TestTokenGenerator_NewToken(t *testing.T) {

	t.Run("should generate token", func(t *testing.T) {
		// given
		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", reName, mock.AnythingOfType("string"))

		tokenGenerator := NewTokenGenerator(tokenLength, tokenCache)

		// when
		newToken, apperr := tokenGenerator.NewToken(reName)

		// then
		require.NoError(t, apperr)

		decoded, err := base64.URLEncoding.DecodeString(newToken)
		require.NoError(t, err)

		assert.Equal(t, tokenLength, len(decoded))
	})

}
