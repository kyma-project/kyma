package endpoints

import (
	"github.com/kyma-project/kyma/components/login-consent/internal/helpers"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (cfg *Config) Consent(w http.ResponseWriter, req *http.Request) {

	challenge, err := helpers.GetLoginChallenge(req)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("Fetching consent request from Hydra")
	consentReq, err := cfg.client.GetConsentRequest(challenge)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if consentReq.Skip {

	}

}
