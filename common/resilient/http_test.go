package resilient_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/resilient"

	retry "github.com/avast/retry-go"
	"github.com/stretchr/testify/assert"
)

type mockHttpClient struct {
	calls        int
	successAfter int
}

func (c *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	if c.calls == c.successAfter {
		return &http.Response{StatusCode: http.StatusTeapot}, nil
	}
	return nil, errors.New("some connection error")
}

func TestHttpClientFirstSuccess(t *testing.T) {
	// given
	mock := &mockHttpClient{successAfter: 1}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))

	// when
	resp, err := wrapped.Get("http://example.com/")

	// then
	assert.Nil(t, err)
	assert.Equal(t, 1, mock.calls)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestHttpClientLateSuccess(t *testing.T) {
	// given
	mock := &mockHttpClient{successAfter: 3}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))

	// when
	resp, err := wrapped.Get("http://example.com/")

	// then
	assert.Nil(t, err)
	assert.Equal(t, 3, mock.calls)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

func TestHttpClientError(t *testing.T) {
	// given
	mock := &mockHttpClient{successAfter: 10}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))

	// when
	_, err := wrapped.Get("http://example.com/")

	// then
	assert.NotNil(t, err)
	assert.Equal(t, 5, mock.calls)
}
