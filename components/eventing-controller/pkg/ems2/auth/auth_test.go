package auth

import (
	"golang.org/x/oauth2"
	"net/http"
	"testing"
)

const (
	// default value in httpTransport implementation
	maxIdleConns        = 100
	maxIdleConnsPerHost = 0
)

func TestAuthenticator(t *testing.T) {

	// authenticate
	authenticator := NewAuthenticator()

	httpClient := authenticator.GetClient().GetHttpClient()

	secTransport, ok := httpClient.Transport.(*oauth2.Transport)
	if !ok {
		t.Errorf("Failed to convert to oauth2 transport")
	}
	httpTransport, ok := secTransport.Base.(*http.Transport)
	if !ok {
		t.Errorf("Failed to convert to HTTP transport")
	}

	if httpTransport.MaxIdleConns != maxIdleConns {
		t.Errorf("HTTP Client Transport MaxIdleConns is misconfigured want: %d but got: %d", maxIdleConns, httpTransport.MaxIdleConns)
	}
	if httpTransport.MaxIdleConnsPerHost != maxIdleConnsPerHost {
		t.Errorf("HTTP Client Transport MaxIdleConnsPerHost is misconfigured want: %d but got: %d", maxIdleConnsPerHost, httpTransport.MaxIdleConnsPerHost)
	}
}
