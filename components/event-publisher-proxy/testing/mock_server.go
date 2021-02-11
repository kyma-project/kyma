package testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Validator is used to validate incoming requests to the mock server
type Validator func(r *http.Request) error

type MockServer struct {
	server               *httptest.Server
	responseTime         time.Duration // server response time
	expiresInSec         int           // token expiry in seconds
	generatedTokensCount int           // generated tokens count
	validator            Validator     // validate the received requests form publishers
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

func WithValidator(validator Validator) MockServerOption {
	return func(m *MockServer) {
		m.validator = validator
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
				if err := m.validateRequest(r); err != nil {
					t.Errorf("request validatation failed with error: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		case eventsWithHTTP400:
			{
				if err := m.validateRequest(r); err != nil {
					t.Errorf("request validatation failed with error: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte("invalid request")); err != nil {
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

func (m *MockServer) validateRequest(r *http.Request) error {
	if m.validator == nil {
		return nil
	}

	return m.validator(r)
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
