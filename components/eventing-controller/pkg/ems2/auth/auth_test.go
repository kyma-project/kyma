package auth

import (
	"fmt"
	"testing"
)

func TestAuth_Login(t *testing.T) {
	authenticator := NewAuthenticator(GetDefaultConfig())
	token, err := authenticator.Authenticate()

	if err != nil {
		t.Fatalf("Failed to authenticate with error: %v", err)
	}

	if token == nil || len(token.Value) == 0 {
		t.Fatal("Received empty token")
	}

	fmt.Printf("%#v\n", token)
}
