package endpoints

import (
	"errors"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/url"
	"time"
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
	loginReq, err := cfg.Client.GetLoginRequest(challenge)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var redirectTo string
	if *loginReq.Skip == true {
		log.Info("Accepting login request")

		body := &hydraAPI.AcceptLoginRequest{
			Remember:    true,
			RememberFor: 30,
			Subject:     nil,
		}

		response, err := cfg.Client.AcceptLoginRequest(challenge, body)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		redirectTo = *response.RedirectTo
	} else {
		log.Info("Showing dex login page")
		log.Infof("Cfg:")
		log.Info(cfg)
		log.Infof("Client:")
		log.Info(cfg.Client)
		log.Infof("Authenticator:")
		log.Info(cfg.Authenticator)
		redirectTo = cfg.Authenticator.clientConfig.AuthCodeURL("state")
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

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
