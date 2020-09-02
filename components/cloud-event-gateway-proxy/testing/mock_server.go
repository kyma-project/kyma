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
	generatedTokensCount int
}

func NewMockServer() *MockServer {
	return &MockServer{generatedTokensCount: 0}
}

func (s *MockServer) Start(t *testing.T, tokenEndpoint, eventsEndpoint string, expiresIn int) {
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case tokenEndpoint:
			{
				s.generatedTokensCount++
				token := fmt.Sprintf("access_token=token-%d&token_type=bearer&expires_in=%d", time.Now().UnixNano(), expiresIn)
				if _, err := w.Write([]byte(token)); err != nil {
					t.Errorf("Failed to write HTTP response")
				}
			}
		case eventsEndpoint:
			{
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			{
				t.Errorf("Mock server supports the following endpoints only: [%s]", tokenEndpoint)
			}
		}
	}))
}

func (s *MockServer) URL() string {
	return s.server.URL
}

func (s *MockServer) GeneratedTokensCount() int {
	return s.generatedTokensCount
}

func (s *MockServer) Close() {
	s.server.Close()
}
