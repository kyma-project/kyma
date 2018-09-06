package proxycache

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	url2 "net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpProxyCache_Create(t *testing.T) {
	t.Run("should create a reverse proxy", func(t *testing.T) {
		// given
		ts := prepareHttpServer(t)
		defer ts.Close()

		proxyCache := NewProxyCache(false, 60)

		url, err := url2.Parse(ts.URL)
		require.NoError(t, err)

		proxy := httputil.NewSingleHostReverseProxy(url)

		// when
		cacheObj := proxyCache.Add("id1", "", "", "", proxy)

		// then
		require.NotNil(t, cacheObj)
		require.NotNil(t, cacheObj.Proxy)

		// when
		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		cacheObj.Proxy.ServeHTTP(rr, req)

		// then
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})
}

func TestHttpProxyCache_Get(t *testing.T) {
	t.Run("should return a proxy cache object", func(t *testing.T) {
		// given
		ts := prepareHttpServer(t)
		defer ts.Close()

		proxyCache := NewProxyCache(false, 60)

		// when
		proxyCache.Add("id1", "", "", "", &httputil.ReverseProxy{})
		cacheObj, _ := proxyCache.Get("id1")

		// then
		require.NotNil(t, cacheObj)
		require.NotNil(t, cacheObj.Proxy)
	})

	t.Run("should return false if id was not found", func(t *testing.T) {
		// given
		proxyCache := NewProxyCache(false, 60)

		// when
		_, found := proxyCache.Get("id1")

		// then
		require.False(t, found)
	})

	t.Run("should return nil if cache expired", func(t *testing.T) {
		// given
		proxyCache := NewProxyCache(false, 2)

		proxyCache.Add("id1", "http://test.url", "", "", &httputil.ReverseProxy{})

		// when
		time.Sleep(3 * time.Second)

		proxy, _ := proxyCache.Get("id1")

		// then
		assert.Nil(t, proxy)
	})
}

func prepareHttpServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.NotContains(t, r.Host, "someurl")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
}
