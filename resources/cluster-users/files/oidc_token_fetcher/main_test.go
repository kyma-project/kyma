package main

import (
	"testing"
)

var oryHydraOauthFqdn = "oauth2.kyma.example.com"
var oidcClientId = "cd13a344-c41e-4d20-b4b9-5654c37288bd"
var oidcClientRedirectUrl = "http://testclient3.example.com"
var email = "admin@kyma.cx"
var password = "1234"

func TestHydraTokenFetcher(t *testing.T) {
	client, err := buildClient()
	if err != nil {
		t.Fatal(err)
	}
	config := oidcConfig{
		email:                 email,
		password:              password,
		oidcClientId:          oidcClientId,
		oidcClientRedirectUrl: oidcClientRedirectUrl,
	}
	h := OidcHydraTestFlow{
		hydraFqdn:  oryHydraOauthFqdn,
		httpClient: client,
		config:     config}
	token, err := h.GetToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Error("token is empty")
	}
}
