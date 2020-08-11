package endpoints

import (
	"errors"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

var challenge string

func (cfg *Config) Login(w http.ResponseWriter, req *http.Request) {
	var err error
	challenge, err = getLoginChallenge(req.URL.Query())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("Fetching login request from Hydra")
	loginReq := hydra.GetLoginRequest(challenge)

	var redirectTo string
	if *loginReq.Skip == true {
		log.Info("Accepting login request")
		response := hydra.AcceptLoginRequest(challenge, hydraAPI.AcceptLoginRequest{
			Remember:    true,
			RememberFor: 30,
			Subject:     nil,
		})
		redirectTo = *response.RedirectTo
	} else {
		log.Info("Showing dex login page")
		redirectTo = cfg.authenticator.clientConfig.AuthCodeURL("state")
	}

	http.Redirect(w, req, redirectTo, http.StatusFound)

}

func getLoginChallenge(query url.Values) (string, error) {
	challenges, ok := query["login_challenge"]
	if !ok || len(challenges[0]) < 1 {
		return "", errors.New("login_challenge not found")
	}

	return challenges[0], nil
}
