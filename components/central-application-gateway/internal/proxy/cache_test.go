package proxy

import (
	"net/http/httputil"
	"testing"

	csrfmocks "github.com/kyma-project/kyma/components/central-application-gateway/internal/csrf/mocks"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/net"
)

func TestCache(t *testing.T) {

	t.Run("should return false if not found", func(t *testing.T) {
		// given
		cache := NewCache(60)

		// when
		cacheEntry, found := cache.Get("app1", "service1", "api1")

		// then
		assert.Nil(t, cacheEntry)
		assert.False(t, found)
	})

	t.Run("should put cache entry", func(t *testing.T) {
		// given
		cache := NewCache(60)

		// when
		authorizationStrategyMock := &mocks.Strategy{}
		csrfTokenStrategy := &csrfmocks.TokenStrategy{}
		url := net.FormatURL("http", "www.example.com", 8080, "")
		proxy := httputil.NewSingleHostReverseProxy(url)

		cacheEntry := cache.Put("app1", "service1", "api1", proxy, authorizationStrategyMock, csrfTokenStrategy)

		// then
		require.NotNil(t, cacheEntry)
		assert.Equal(t, proxy, cacheEntry.Proxy)
		assert.Equal(t, authorizationStrategyMock, cacheEntry.AuthorizationStrategy.actualStrategy)
		assert.Equal(t, csrfTokenStrategy, cacheEntry.CSRFTokenStrategy)

		// when
		cacheEntry, found := cache.Get("app1", "service1", "api1")

		// then
		require.NotNil(t, cacheEntry)
		assert.True(t, found)
		assert.Equal(t, proxy, cacheEntry.Proxy)
		assert.Equal(t, authorizationStrategyMock, cacheEntry.AuthorizationStrategy.actualStrategy)
		assert.Equal(t, csrfTokenStrategy, cacheEntry.CSRFTokenStrategy)
	})
}
