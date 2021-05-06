package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/icza/session"
	"github.com/kyma-project/kyma/components/busola-migrator/pkg/rand"

	"github.com/pkg/errors"
)

const (
	oauthStateLen              = 16
	oauthStateSessionAttribute = "oauthstate"
)

func (a App) HandleXSUAAMigrate(w http.ResponseWriter, r *http.Request) {
	state, err := rand.Hex(oauthStateLen)
	if err != nil {
		log.Println(errors.Wrap(err, "while generating random state string"))
		http.Error(w, "Could not generate state string", http.StatusInternalServerError)
		return
	}

	sess := session.Get(r)
	if sess == nil {
		sess = session.NewSession()
		session.Add(sess, w)
	}

	sess.SetAttr(oauthStateSessionAttribute, state)

	url, err := a.uaaClient.GetAuthorizationEndpointWithParams(a.uaaOIDConfig.AuthorizationEndpoint, state)
	if err != nil {
		log.Println(errors.Wrap(err, "while getting UAA authorization endpoint with encoded params"))
		http.Error(w, "Invalid UAA authorization request params", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (a App) HandleXSUAACallback(w http.ResponseWriter, r *http.Request) {
	redirectHost := strings.Replace(r.Host, "dex", "console", 1)
	if strings.HasPrefix(r.Host, "dex") {
		http.Redirect(w, r, fmt.Sprintf("https://%s/callback?%s", redirectHost, r.URL.RawQuery), http.StatusFound)
		return
	}

	sess := session.Get(r)
	if sess == nil {
		log.Println(errors.New("Session not found"))
		http.Error(w, "Session not found", http.StatusInternalServerError)
		return
	}

	// retrieve stored state
	storedState, ok := sess.Attr(oauthStateSessionAttribute).(string)
	if !ok {
		log.Println(errors.New("Invalid OAuth state"))
		http.Error(w, "Invalid OAuth state", http.StatusInternalServerError)
		return
	}

	// check if state returned by oauth server matches stored state
	if r.URL.Query().Get("state") != storedState {
		log.Println(errors.New("Invalid OAuth state"))
		http.Error(w, "Invalid OAuth state", http.StatusInternalServerError)
		return
	}

	// remove stored state string to prevent request repeats
	sess.SetAttr(oauthStateSessionAttribute, nil)

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

	http.Redirect(w, r, fmt.Sprintf("https://%s/info/success.html", redirectHost), http.StatusFound)
}
