package tokencache

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	cachedClientID = "cachedClientID"
	cachedToken    = "cachedToken"
)

func TestTokenCache(t *testing.T) {
	t.Run("should add and retrieve the cachedToken from the cache", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(cachedClientID, cachedToken, 3600)

		// when
		token, found := tokenCache.Get(cachedClientID)

		// then
		assert.Equal(t, true, found)
		assert.Equal(t, cachedToken, token)

	})

	t.Run("should return false if cachedToken was not found", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()

		// when
		token, found := tokenCache.Get(cachedClientID)

		// then
		assert.Equal(t, false, found)
		assert.Equal(t, "", token)
	})

	t.Run("should return false if cachedToken expired", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(cachedClientID, cachedToken, 3)

		time.Sleep(3 * time.Second)

		// when
		token, found := tokenCache.Get(cachedClientID)

		// then
		assert.Equal(t, false, found)
		assert.Equal(t, "", token)
	})

	t.Run("should remove token from the cache", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(cachedClientID, cachedToken, 3600)
		tokenCache.Remove(cachedClientID)

		// when
		token, found := tokenCache.Get(cachedClientID)

		// then
		assert.Equal(t, false, found)
		assert.Equal(t, "", token)
	})
}
