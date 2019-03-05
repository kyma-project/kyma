package csrf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	itemId       = "someEndpointURL"
	cachedToken  = "someToken"
	cachedCookie = "someCookie"
)

func TestTokenCache(t *testing.T) {

	resp := &Response{
		csrfTokenEndpointUrl: cachedToken,
		cookie:               cachedCookie,
	}

	t.Run("should add and retrieve the response from the cache", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(itemId, resp)

		// when
		response, found := tokenCache.Get(itemId)

		// then
		assert.Equal(t, true, found)
		assert.Equal(t, cachedToken, response.csrfTokenEndpointUrl)
		assert.Equal(t, cachedCookie, response.cookie)
	})

	t.Run("should return false if the response was not found", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()

		// when
		resp, found := tokenCache.Get(itemId)

		// then
		assert.Equal(t, false, found)
		assert.Nil(t, resp)
	})

	t.Run("should remove a response from the cache", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(itemId, resp)
		tokenCache.Remove(itemId)

		// when
		resp, found := tokenCache.Get(itemId)

		// then
		assert.Equal(t, false, found)
		assert.Nil(t, resp)
	})

}
