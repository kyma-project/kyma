package endpoints

import (
	"github.com/kyma-project/kyma/components/login-consent/internal/helpers"
	hydraAPI "github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (cfg *Config) Login(w http.ResponseWriter, req *http.Request) {
	var err error
	challenge, err = helpers.GetLoginChallenge(req)
	if err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusBadRequest)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusBadRequest)
		return
	}

	log.Info("Fetching login request from Hydra")
	loginReq, err := cfg.client.GetLoginRequest(challenge)
	if err != nil {
		redirect, err := cfg.rejectLoginRequest(err, http.StatusBadRequest)
		if err != nil {
			log.Errorf("failed to reject the login request: %s", err)
			return
		}
		http.Redirect(w, req, redirect, http.StatusBadRequest)
		return
	}

	var redirectTo string
	if *loginReq.Skip {
		log.Info("accepting login request...")

		body := &hydraAPI.AcceptLoginRequest{
			Remember:    false,
			RememberFor: 30,
			Subject:     nil,
		}

		response, err := cfg.client.AcceptLoginRequest(challenge, body)
		if err != nil {
			redirect, err := cfg.rejectLoginRequest(err, http.StatusBadRequest)
			if err != nil {
				log.Errorf("failed to reject the login request: %s", err)
				return
			}
			http.Redirect(w, req, redirect, http.StatusInternalServerError)
			return
		}

		redirectTo = *response.RedirectTo
	} else {
		redirectTo = cfg.authenticator.clientConfig.AuthCodeURL(state)
	}

	log.Infof("redirecting to: %s", redirectTo)
	http.Redirect(w, req, redirectTo, http.StatusFound)

}
