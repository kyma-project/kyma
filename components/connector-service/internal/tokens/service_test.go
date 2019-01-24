package tokens

import (
	"bytes"
	"encoding/json"
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

type dummySerializable struct {
	Value []byte
	Error error
}

func (params dummySerializable) ToJSON() ([]byte, error) {
	return params.Value, params.Error
}

func TestService_Save(t *testing.T) {
	t.Run("should trigger Put metod on token store", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Put", token, payload)
		tokenGenerator := func() (string, apperrors.AppError) { return token, nil }

		tokenService := NewTokenService(tokenCache, tokenGenerator)

		generatedToken, err := tokenService.Save(serializable)

		require.NoError(t, err)
		assert.Equal(t, token, generatedToken)
	})

	t.Run("should return error when failed on token serialization", func(t *testing.T) {

		serializable := dummySerializable{
			Value: nil,
			Error: errors.New("error"),
		}
		tokenService := NewTokenService(nil, nil)

		_, err := tokenService.Save(serializable)

		require.Error(t, err)
	})

	t.Run("should return error when generator fails to generate token", func(t *testing.T) {

		serializable := dummySerializable{
			Value: []byte(payload),
			Error: nil,
		}

		tokenGenerator := func() (string, apperrors.AppError) { return "", apperrors.Internal("error") }
		tokenService := NewTokenService(nil, tokenGenerator)

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
		tokenCache.On("Put", newToken, payload)
		tokenGenerator := func() (string, apperrors.AppError) { return newToken, nil }

		tokenService := NewTokenService(tokenCache, tokenGenerator)

		generatedToken, err := tokenService.Replace(token, serializable)

		require.NoError(t, err)
		assert.Equal(t, newToken, generatedToken)
	})
}

func TestTokenService_Resolve(t *testing.T) {

	dummyString := "data"
	encodedData := string(compact([]byte("{\"data\":\"data\"}")))

	type data struct {
		Data string `json:"data"`
	}

	dummyData := data{Data: dummyString}

	t.Run("shoud resolve token", func(t *testing.T) {
		// given
		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Get", token).Return(encodedData, true)

		var destination data

		tokenService := NewTokenService(tokenCache, nil)

		// when
		err := tokenService.Resolve(token, &destination)

		// then
		require.NoError(t, err)
		assert.Equal(t, dummyData.Data, destination.Data)
	})

	// TODO - more tests
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
