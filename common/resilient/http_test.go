package resilient

import (
	"errors"
	"github.com/avast/retry-go"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type mockHttpClientFirstSuccess struct{}
func (mockHttpClientFirstSuccess) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: http.StatusTeapot}, nil
}

func TestHttpClientFirstSuccess(t *testing.T) {
	mock := &mockHttpClientFirstSuccess{}
	wrapped := WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	resp, err := wrapped.Get("http://example.com/")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}

type mockHttpClientLateSuccess struct{
	mockHttpClientFirstSuccess
	*mockHttpClientError
	successAfter int
}
func (c *mockHttpClientLateSuccess) Do(req *http.Request) (*http.Response, error) {
	if c.calls == c.successAfter {
		return c.mockHttpClientFirstSuccess.Do(req)
	}
	return c.mockHttpClientError.Do(req)
}

func TestHttpClientLateSuccess(t *testing.T) {
	mock := &mockHttpClientLateSuccess{successAfter: 3, mockHttpClientError: &mockHttpClientError{}}
	wrapped := WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	resp, err := wrapped.Get("http://example.com/")
	assert.Nil(t, err)
	assert.Equal(t, 3, mock.calls)
	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
}


type mockHttpClientError struct{
	calls int
}
func (c *mockHttpClientError) Do(req *http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("some connection error")
}

func TestHttpClientError(t *testing.T) {
	mock := &mockHttpClientError{}
	wrapped := WrapHttpClient(mock, retry.Delay(time.Millisecond), retry.Attempts(5))
	_, err := wrapped.Get("http://example.com/")
	assert.NotNil(t, err)
	assert.Equal(t, 5, mock.calls)
}