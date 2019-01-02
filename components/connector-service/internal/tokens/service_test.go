package tokens_test

import (
	"encoding/base64"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tokenLength = 10
	appName     = "appName"
	group       = "group"
	tenant      = "tenant"
)

func TestTokenGenerator_NewToken(t *testing.T) {

	t.Run("should generate token", func(t *testing.T) {
		// given
		tokenData := &tokens.TokenData{
			Group:  group,
			Tenant: tenant,
		}

		tokenCache := &mocks.Cache{}
		tokenCache.On("Put", appName, tokenData)

		tokenService := tokens.NewTokenService(tokenLength, tokenCache)

		// when
		newToken, apperr := tokenService.CreateToken(appName, tokenData)

		// then
		require.NoError(t, apperr)

		decoded, err := base64.URLEncoding.DecodeString(newToken)
		require.NoError(t, err)

		assert.Equal(t, tokenLength, len(decoded))
	})

}
