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
	t.Run("should trigger Put method on token store", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", token, payload, tokenTTL)
		tokenGenerator := func() (string, apperrors.AppError) { return token, nil }

		tokenCreator := NewTokenCreator(tokenTTL, tokenCache, tokenGenerator)

		generatedToken, err := tokenCreator.Save(serializable)

		require.NoError(t, err)
		assert.Equal(t, token, generatedToken)
	})

	t.Run("should return error when failed on token serialization", func(t *testing.T) {

		serializable := dummySerializable{
			Value: nil,
			Error: errors.New("error"),
		}
		tokenCreator := NewTokenCreator(tokenTTL, nil, nil)

		_, err := tokenCreator.Save(serializable)

		require.Error(t, err)
	})

	t.Run("should return error when generator fails to generate token", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenGenerator := func() (string, apperrors.AppError) { return "", apperrors.Internal("error") }
		tokenCreator := NewTokenCreator(tokenTTL, nil, tokenGenerator)

		_, err := tokenCreator.Save(serializable)

		require.Error(t, err)
	})
}
