package endpoints

import (
	"github.com/coreos/go-oidc"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (cfg *Config) Callback(w http.ResponseWriter, req *http.Request) {
	log.Info("Checking state match")
	if req.URL.Query().Get("state") != "state" {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	log.Info("Exchanging code for token")
	token, err := cfg.authenticator.clientConfig.Exchange(cfg.authenticator.ctx, req.URL.Query().Get("code"))
	if err != nil {
		log.Printf("no token found: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	log.Infof("Raw: %s", rawIDToken)

	oidcConfig := &oidc.Config{
		ClientID: cfg.authenticator.clientConfig.ClientID, //TODO provide proper data here
	}

	log.Info("Verifying ID Token")
	idToken, err := cfg.authenticator.provider.Verifier(oidcConfig).Verify(cfg.authenticator.ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof("Ready: %s", idToken)

	//resp := struct {
	//	OAuth2Token   *oauth2.Token
	//	IDTokenClaims *json.RawMessage
	//}{token, new(json.RawMessage)}
	//
	//log.Infof("Verifier response: %s", resp)
	//if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	acceptLoginRequest := hydraAPI.AcceptLoginRequest{
		Context:     idToken,
		Remember:    idToken.Expiry.IsZero(),
		RememberFor: 3600,
		Subject:     &idToken.Subject,
	}

	hydraResp := hydra.AcceptLoginRequest(challenge, acceptLoginRequest)

	http.Redirect(w, req, *hydraResp.RedirectTo, http.StatusFound)

	//data, err := json.MarshalIndent(resp, "", "    ")
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//TODO: say hello to hydra instead of writing the token down
	//w.Write([]byte("ok"))
}
