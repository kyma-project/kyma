package tokens

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	token   = "abc"
	payload = "data"
)

type dummyTokenParams struct {
	Value []byte
	Error error
}

func (params dummyTokenParams) ToJSON() ([]byte, error) {
	return params.Value, params.Error
}

func TestService_Save(t *testing.T) {
	t.Run("should trigger Put metod on token store", func(t *testing.T) {

		dummyParams := dummyTokenParams{
			Value: []byte(payload),
			Error: nil,
		}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", payload, token)
		tokenGenerator := func(length int) (string, apperrors.AppError) { return token, nil }
		tokenService := NewTokenService(3, tokenCache, tokenGenerator)

		generatedToken, err := tokenService.Save(dummyParams)

		require.NoError(t, err)
		assert.Equal(t, token, generatedToken)
	})

	t.Run("should return error when failed on token serialization", func(t *testing.T) {

		dummyParams := dummyTokenParams{
			Value: nil,
			Error: errors.New("error"),
		}
		tokenService := NewTokenService(3, nil, nil)

		_, err := tokenService.Save(dummyParams)

		require.Error(t, err)
	})

	t.Run("should return error when generator fails to generate token", func(t *testing.T) {

		dummyParams := dummyTokenParams{
			Value: []byte(payload),
			Error: nil,
		}

		tokenGenerator := func(length int) (string, apperrors.AppError) { return "", apperrors.Internal("error") }
		tokenService := NewTokenService(3, nil, tokenGenerator)

		_, err := tokenService.Save(dummyParams)

		require.Error(t, err)
	})
}
