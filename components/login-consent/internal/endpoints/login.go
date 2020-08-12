package endpoints

import (
	"github.com/kyma-project/kyma/components/login-consent/internal/helpers"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"time"
)

var challenge string
var state = generateRandomString(16)

func (cfg *Config) Login(w http.ResponseWriter, req *http.Request) {
	var err error
	challenge, err = helpers.GetLoginChallenge(req)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("Fetching login request from Hydra")
	loginReq, err := cfg.client.GetLoginRequest(challenge)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var redirectTo string
	if *loginReq.Skip {
		log.Info("Accepting login request")

		body := &hydraAPI.AcceptLoginRequest{
			Remember:    true,
			RememberFor: 30,
			Subject:     nil,
		}

		response, err := cfg.client.AcceptLoginRequest(challenge, body)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		redirectTo = *response.RedirectTo
	} else {

		redirectTo = cfg.authenticator.clientConfig.AuthCodeURL(state)
	}

	http.Redirect(w, req, redirectTo, http.StatusFound)

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
