package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func (a App) HandleXSUAAMigrate(w http.ResponseWriter, r *http.Request) {
	url, err := a.uaaClient.GetAuthorizationEndpointWithParams(a.uaaOIDConfig.AuthorizationEndpoint)
	if err != nil {
		log.Println(errors.Wrap(err, "while getting UAA authorization endpoint with encoded params"))
		http.Error(w, "Invalid UAA authorization request params", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (a App) HandleXSUAACallback(w http.ResponseWriter, r *http.Request) {
	body, err := a.uaaClient.GetToken(a.uaaOIDConfig.TokenEndpoint, r.URL.Query().Get("code"))
	if err != nil {
		log.Println(errors.Wrap(err, "while getting UAA token"))
		http.Error(w, "Error while getting UAA token", http.StatusInternalServerError)
		return
	}

	token, ok := body["access_token"].(string)
	if !ok {
		log.Println(errors.New("UAA token response did not contain access token"))
		http.Error(w, "UAA token response did not contain access token", http.StatusInternalServerError)
		return
	}

	parsed, err := a.jwtService.ParseAndVerify(token, a.uaaOIDConfig.JWKSURI)
	if err != nil {
		log.Println(errors.Wrap(err, "while verifying UAA access token"))
		http.Error(w, "UAA access token could not be verified", http.StatusInternalServerError)
		return
	}

	user, err := a.jwtService.GetUser(parsed)
	if err != nil {
		log.Println(errors.Wrap(err, "while getting user permissions from access token"))
		http.Error(w, "UAA access token did not contain valid scopes claim", http.StatusInternalServerError)
		return
	}

	err = a.k8sClient.EnsureUserPermissions(user)
	if err != nil {
		log.Println(errors.Wrap(err, "while creating Role Bindings for user"))
		http.Error(w, "Error while creating Role Bindings for user", http.StatusInternalServerError)
		return
	}

	redirectHost := strings.Replace(r.Host, "dex", "console", 1)
	http.Redirect(w, r, fmt.Sprintf("https://%s/info/success.html", redirectHost), http.StatusFound)
}
