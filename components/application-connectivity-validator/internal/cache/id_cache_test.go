package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("should return not found if cache is empty", func(t *testing.T) {
		// given
		applicationName := "app"
		expirationMinutes := 1
		cleanupMinutes := 2

		cache := NewCache(expirationMinutes, cleanupMinutes)

		// when
		_, found := cache.GetClientIDs(applicationName)

		// then
		assert.False(t, found)
	})

	t.Run("should return clientIDs if they exist in cache", func(t *testing.T) {
		// given
		applicationName := "app"
		expirationMinutes := 1
		cleanupMinutes := 2
		clientIDs := []string{
			"someID1",
			"someID2",
		}

		cache := NewCache(expirationMinutes, cleanupMinutes)

		cache.SetClientIDs(applicationName, clientIDs)

		// when
		returnedClientIDs, found := cache.GetClientIDs(applicationName)

		// then
		require.True(t, found)
		assert.Equal(t, clientIDs, returnedClientIDs)
	})

	t.Run("should return not found if cache does not contain items for provided key", func(t *testing.T) {
		// given
		applicationName := "app"
		anotherApplicationName := "anotherapp"
		expirationMinutes := 1
		cleanupMinutes := 2
		clientIDs := []string{
			"someID1",
			"someID2",
		}

		cache := NewCache(expirationMinutes, cleanupMinutes)

		cache.SetClientIDs(applicationName, clientIDs)

		// when
		_, found := cache.GetClientIDs(anotherApplicationName)

		// then
		assert.False(t, found)
	})
}
