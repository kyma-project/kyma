package tokens

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenService_Resolve(t *testing.T) {

	type data struct {
		Data string `json:"data"`
	}

	t.Run("should resolve token", func(t *testing.T) {
		// given
		dummyString := "data"
		encodedData := string(compact([]byte("{\"data\":\"data\"}")))
		dummyData := data{Data: dummyString}

		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Get", token).Return(encodedData, true)

		var destination data

		tokenManager := NewTokenManager(tokenCache)

		// when
		err := tokenManager.Resolve(token, &destination)

		// then
		require.NoError(t, err)
		assert.Equal(t, dummyData.Data, destination.Data)
	})

	t.Run("should return error when token not found", func(t *testing.T) {
		// given
		tokenCache := &mocks.TokenCache{}
		tokenCache.On("Get", token).Return("", false)

		var destination data

		tokenManager := NewTokenManager(tokenCache)

		// when
		err := tokenManager.Resolve(token, &destination)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeNotFound, err.Code())
	})
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
