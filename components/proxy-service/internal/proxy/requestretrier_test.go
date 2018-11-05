package proxy

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestRequestRetrier_CheckResponse(t *testing.T) {
	t.Run("should return response if code is different than 403", func(t *testing.T) {
		// given
		rr := newRequestRetrier("", &http.Request{}, &CacheEntry{}, 10)
		response := &http.Response{StatusCode: 200}

		// when
		err := rr.CheckResponse(response)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not retry if flag is already set", func(t *testing.T) {
		// given
		rr := newRequestRetrier("", &http.Request{}, &CacheEntry{}, 10)
		rr.retried = true
		response := &http.Response{StatusCode: 403}

		// when
		err := rr.CheckResponse(response)

		// then
		assert.NoError(t, err)
	})
}
