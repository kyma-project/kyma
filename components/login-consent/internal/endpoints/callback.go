package endpoints

import (
	"errors"
	"net/http"

	"github.com/coreos/go-oidc"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
)

func (cfg *Config) Callback(w http.ResponseWriter, req *http.Request) {

	log.Infof("DEBUG: Callback endpoint hit with req.URL: %s", req.URL.String())
	log.Info("checking state match")

	if req.URL.Query().Get("state") != state {
		redirect, err := cfg.rejectLoginRequest(errors.New("state doesn't match"), http.StatusUnauthorized)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusUnauthorized)
		return
	}

	log.Info("exchanging code for token")
	token, err := cfg.authenticator.tokenSupport.Exchange(cfg.authenticator.ctx, req.URL.Query().Get("code"))
	if err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusUnauthorized)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		redirect, err := cfg.rejectLoginRequest(errors.New("no id_token field in oauth2 token"), http.StatusInternalServerError)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusInternalServerError)
		return
	}
	log.Infof("Raw: %s", rawIDToken)

	oidcConfig := &oidc.Config{
		ClientID: cfg.authenticator.tokenSupport.ClientID(),
	}

	log.Info("verifying ID Token...")
	idToken, err := cfg.authenticator.tokenSupport.Verify(oidcConfig, cfg.authenticator.ctx, rawIDToken)
	if err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusInternalServerError)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusInternalServerError)
		return
	}

	log.Infof("token verified")

	var claims struct {
		Email         string   `json:"email"`
		EmailVerified bool     `json:"email_verified"`
		Name          string   `json:"name"`
		Groups        []string `json:"groups,omitempty"`
	}

	if err := idToken.Claims(&claims); err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusUnauthorized)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusUnauthorized)
		return
	}
	acceptLoginRequest := &hydraAPI.AcceptLoginRequest{
		Context:     claims,
		Remember:    false, //TODO: change this later
		RememberFor: 3600,
		Subject:     &idToken.Subject,
	}

	log.Infof("accepting login request")
	hydraResp, err := cfg.hydraClient.AcceptLoginRequest(challenge, acceptLoginRequest)
	if err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusUnauthorized)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusUnauthorized)
		return
	}

	log.Infof("Redirecting to: %s", *hydraResp.RedirectTo)
	http.Redirect(w, req, *hydraResp.RedirectTo, http.StatusFound)
}
