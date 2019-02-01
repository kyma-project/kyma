package tokens

import (
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	token    = "abc"
	payload  = "data"
	tokenTTL = time.Duration(5) * time.Minute
)

type dummySerializable struct {
	Value []byte
	Error error
}

func (params dummySerializable) ToJSON() ([]byte, error) {
	return params.Value, params.Error
}

func TestManager_Save(t *testing.T) {
	t.Run("should trigger Put metod on token store", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", token, payload, tokenTTL)
		tokenGenerator := func() (string, apperrors.AppError) { return token, nil }

		tokenManager := NewTokenManager(tokenTTL, tokenCache, tokenGenerator)

		generatedToken, err := tokenManager.Save(serializable)

		require.NoError(t, err)
		assert.Equal(t, token, generatedToken)
	})

	t.Run("should return error when failed on token serialization", func(t *testing.T) {

		serializable := dummySerializable{
			Value: nil,
			Error: errors.New("error"),
		}
		tokenService := NewTokenManager(tokenTTL, nil, nil)

		_, err := tokenService.Save(serializable)

		require.Error(t, err)
	})

	t.Run("should return error when generator fails to generate token", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenGenerator := func() (string, apperrors.AppError) { return "", apperrors.Internal("error") }
		tokenService := NewTokenManager(tokenTTL, nil, tokenGenerator)

		_, err := tokenService.Save(serializable)

		require.Error(t, err)
	})
}

func TestTokenService_Replace(t *testing.T) {
	newToken := "newToken"

	t.Run("should delete token before saving new one", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Delete", token)
		tokenCache.On("Put", newToken, payload, tokenTTL)
		tokenGenerator := func() (string, apperrors.AppError) { return newToken, nil }

		tokenService := NewTokenManager(tokenTTL, tokenCache, tokenGenerator)

		generatedToken, err := tokenService.Replace(token, serializable)

		require.NoError(t, err)
		assert.Equal(t, newToken, generatedToken)
	})
}
