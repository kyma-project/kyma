package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestRequestRetrier_CheckResponse(t *testing.T) {
	t.Run("should return response if code is different than 403", func(t *testing.T) {
		// given
		rr := newRequestRetrier("", &proxy{}, &http.Request{})
		response := &http.Response{StatusCode: 200}

		// when
		err := rr.CheckResponse(response)

		// then
		assert.NoError(t, err)
	})
}
