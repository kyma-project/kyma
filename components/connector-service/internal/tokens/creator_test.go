package tokens

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	token    = "abc"
	tokenTTL = time.Duration(5) * time.Minute
)

type notSerializable string

func (_ notSerializable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("error")
}

func TestManager_Save(t *testing.T) {
	serializable := &clientcontext.ApplicationContext{
		Application: "app",
	}
	t.Run("should trigger Put method on token store", func(t *testing.T) {

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", token, mock.AnythingOfType("string"), tokenTTL)
		tokenGenerator := func() (string, apperrors.AppError) { return token, nil }

		tokenCreator := NewTokenCreator(tokenTTL, tokenCache, tokenGenerator)

		generatedToken, err := tokenCreator.Save(serializable)

		require.NoError(t, err)
		assert.Equal(t, token, generatedToken)
	})

	t.Run("should return error when failed on token serialization", func(t *testing.T) {
		tokenCreator := NewTokenCreator(tokenTTL, nil, nil)

		_, err := tokenCreator.Save(notSerializable("abc"))

		require.Error(t, err)
	})

	t.Run("should return error when generator fails to generate token", func(t *testing.T) {
		tokenGenerator := func() (string, apperrors.AppError) { return "", apperrors.Internal("error") }
		tokenCreator := NewTokenCreator(tokenTTL, nil, tokenGenerator)

		_, err := tokenCreator.Save(serializable)

		require.Error(t, err)
	})
}
