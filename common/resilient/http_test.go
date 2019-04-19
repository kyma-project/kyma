package resilient_test

import (
	"errors"
	"github.com/kyma-project/kyma/common/resilient"
	"net/http"
	"testing"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/stretchr/testify/assert"
)

type mockHttpClientFirstSuccess struct{}

func (mockHttpClientFirstSuccess) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusTeapot}, nil
}

func TestHttpClientFirstSuccess(t *testing.T) {
	mock := &mockHttpClientFirstSuccess{}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	resp, err := wrapped.Get("http://example.com/")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

type mockHttpClientLateSuccess struct {
	calls        int
	successAfter int
}

func (c *mockHttpClientLateSuccess) Do(req *http.Request) (*http.Response, error) {
	if c.calls == c.successAfter {
		return &http.Response{StatusCode: http.StatusTeapot}, nil
	}
	c.calls++
	return nil, errors.New("some connection error")
}

func TestHttpClientLateSuccess(t *testing.T) {
	mock := &mockHttpClientLateSuccess{successAfter: 3}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	resp, err := wrapped.Get("http://example.com/")
	assert.Nil(t, err)
	assert.Equal(t, 3, mock.calls)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

type mockHttpClientError struct {
	calls int
}

func (c *mockHttpClientError) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("some connection error")
}

func TestHttpClientError(t *testing.T) {
	mock := &mockHttpClientError{}
	wrapped := resilient.WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	_, err := wrapped.Get("http://example.com/")
	assert.NotNil(t, err)
	assert.Equal(t, 5, mock.calls)
}
