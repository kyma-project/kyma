package testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockServer struct {
	server               *httptest.Server
	responseTime         time.Duration // server response time
	expiresInSec         int           // token expiry in seconds
	generatedTokensCount int           // generated tokens count
}

func NewMockServer(opts ...MockServerOption) *MockServer {
	mockServer := &MockServer{expiresInSec: 0, generatedTokensCount: 0, responseTime: 0}
	for _, opt := range opts {
		opt(mockServer)
	}
	return mockServer
}

type MockServerOption func(m *MockServer)

func WithExpiresIn(expiresIn int) MockServerOption {
	return func(m *MockServer) {
		m.expiresInSec = expiresIn
	}
}

func WithResponseTime(responseTime time.Duration) MockServerOption {
	return func(m *MockServer) {
		m.responseTime = responseTime
	}
}

func (m *MockServer) Start(t *testing.T, tokenEndpoint, eventsEndpoint, eventsWithHTTP400 string) {
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(m.responseTime)

		switch r.URL.String() {
		case tokenEndpoint:
			{
				m.generatedTokensCount++
				token := fmt.Sprintf("access_token=token-%d&token_type=bearer&expires_in=%d", time.Now().UnixNano(), m.expiresInSec)
				if _, err := w.Write([]byte(token)); err != nil {
					t.Errorf("failed to write HTTP response")
				}
			}
		case eventsEndpoint:
			{
				w.WriteHeader(http.StatusNoContent)
			}
		case eventsWithHTTP400:
			{
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("invalid request"))
				if err != nil {
					t.Errorf("failed to write message: %v", err)
				}
			}
		default:
			{
				t.Errorf("mock server supports the following endpoints only: [%s]", tokenEndpoint)
			}
		}
	}))
}

func (m *MockServer) URL() string {
	return m.server.URL
}

func (m *MockServer) GeneratedTokensCount() int {
	return m.generatedTokensCount
}

func (m *MockServer) Close() {
	m.server.Close()
}
