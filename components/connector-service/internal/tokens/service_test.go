package tokens_test

import (
	"testing"

	"github.com/patrickmn/go-cache"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"

	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tokenLength = 10
	identifier  = "identifier"
	group       = "group"
	tenant      = "tenant"
	token       = "token"
)

var (
	initialTokenData = &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
	}

	completeTokenData = &tokens.TokenData{
		Group:  group,
		Tenant: tenant,
		Token:  token,
	}
)

func TestTokenService_CreateToken(t *testing.T) {

	t.Run("should create and save new app token", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}
		appTokenCache.On("Set", identifier, completeTokenData, cache.DefaultExpiration)

		clusterTokenCache := &mocks.Cache{}

		tokenService := tokens.NewTokenService(tokenLength, tokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		newToken, apperr := tokenService.CreateAppToken(identifier, initialTokenData)

		// then
		require.NoError(t, apperr)
		assert.Equal(t, token, newToken)
	})

	t.Run("should create and save new cluster token", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}

		clusterTokenCache := &mocks.Cache{}
		clusterTokenCache.On("Set", identifier, token, cache.DefaultExpiration)

		tokenService := tokens.NewTokenService(tokenLength, tokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		newToken, apperr := tokenService.CreateClusterToken(identifier)

		// then
		require.NoError(t, apperr)
		assert.Equal(t, token, newToken)
	})

	t.Run("should return error when failed to generate app token", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}
		clusterTokenCache := &mocks.Cache{}

		tokenService := tokens.NewTokenService(tokenLength, failTokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		newToken, apperr := tokenService.CreateAppToken(identifier, initialTokenData)

		// then
		require.Error(t, apperr)
		assert.Equal(t, "", newToken)
		assert.Equal(t, apperrors.CodeInternal, apperr.Code())
	})

	t.Run("should return error when failed to generate cluster token", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}
		clusterTokenCache := &mocks.Cache{}

		tokenService := tokens.NewTokenService(tokenLength, failTokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		newToken, apperr := tokenService.CreateClusterToken(identifier)

		// then
		require.Error(t, apperr)
		assert.Equal(t, "", newToken)
		assert.Equal(t, apperrors.CodeInternal, apperr.Code())
	})
}

func TestTokenService_GetToken(t *testing.T) {

	t.Run("should get app token data", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}
		appTokenCache.On("Get", identifier).Return(completeTokenData, true)

		clusterTokenCache := &mocks.Cache{}

		tokenService := tokens.NewTokenService(tokenLength, tokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		tokenData, found := tokenService.GetAppToken(identifier)

		// then
		require.True(t, found)
		assert.Equal(t, completeTokenData, tokenData)
	})

	t.Run("should get cluster token data", func(t *testing.T) {
		// given
		appTokenCache := &mocks.Cache{}

		clusterTokenCache := &mocks.Cache{}
		clusterTokenCache.On("Get", identifier).Return(token, true)

		tokenService := tokens.NewTokenService(tokenLength, tokenGenFunc, appTokenCache, clusterTokenCache)

		// when
		cachedToken, found := tokenService.GetClusterToken(identifier)

		// then
		require.True(t, found)
		assert.Equal(t, token, cachedToken)
	})
}

func tokenGenFunc(l int) (string, apperrors.AppError) {
	return token, nil
}

func failTokenGenFunc(l int) (string, apperrors.AppError) {
	return "", apperrors.Internal("error")
}
