package client

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf"
	"github.com/stretchr/testify/assert"
)

const (
	itemId           = "someEndpointURL"
	cachedToken      = "someToken"
	cachedCookieName = "someCookie"
)

func TestTokenCache(t *testing.T) {

	testCookie := http.Cookie{Name: cachedCookieName}

	resp := &csrf.Response{
		CSRFToken: cachedToken,
		Cookies:   []*http.Cookie{&testCookie},
	}

	t.Run("should add and retrieve the response from the cache", func(t *testing.T) {
		// given
		tokenCache := NewTokenCache()
		tokenCache.Add(itemId, resp)

		// when
		response, found := tokenCache.Get(itemId)

		// then
		assert.Equal(t, true, found)
		assert.Equal(t, cachedToken, response.CSRFToken)
		assert.Equal(t, cachedCookieName, response.Cookies[0].Name)
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
