package auth

import (
	"net/http"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"golang.org/x/oauth2"
)

const (
	// default value in httpTransport implementation
	maxIdleConns        = 100
	maxIdleConnsPerHost = 0
)

func TestAuthenticator(t *testing.T) {
	if err := os.Setenv("CLIENT_ID", "foo"); err != nil {
		t.Errorf("error while setting env var CLIENT_ID")
	}
	if err := os.Setenv("CLIENT_SECRET", "foo"); err != nil {
		t.Errorf("error while setting env var CLIENT_SECRET")
	}
	if err := os.Setenv("TOKEN_ENDPOINT", "foo"); err != nil {
		t.Errorf("error while setting env var TOKEN_ENDPOINT")
	}
	cfg := env.Config{}
	// authenticate
	authenticator := NewAuthenticator(cfg)

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
