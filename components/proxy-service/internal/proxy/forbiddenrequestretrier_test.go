package proxy

import (
	"github.com/kyma-project/kyma/components/proxy-service/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestForbiddenRequestRetrier_CheckResponse(t *testing.T) {
	updateCacheEntryFunction := func(id string) (*CacheEntry, apperrors.AppError) {
		return nil, nil
	}

	t.Run("should return response if code is different than 403", func(t *testing.T) {
		// given
		rr := newForbiddenRequestRetrier("", &http.Request{}, 10, updateCacheEntryFunction)
		response := &http.Response{StatusCode: 200}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not retry if flag is already set", func(t *testing.T) {
		// given
		rr := newForbiddenRequestRetrier("", &http.Request{}, 10, updateCacheEntryFunction)
		rr.retried = true
		response := &http.Response{StatusCode: 403}

		// when
		err := rr.RetryIfFailedToAuthorize(response)

		// then
		assert.NoError(t, err)
	})
}
