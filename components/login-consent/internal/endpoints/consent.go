package endpoints

import (
	"github.com/kyma-project/kyma/components/login-consent/internal/helpers"
	"github.com/ory/hydra-client-go/models"
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

	if challenge == "" {
		w.Write([]byte("expected a consent challenge to be set but received none."))
		w.WriteHeader(http.StatusBadRequest)
	}

	log.Info("Fetching consent request from Hydra")
	consentReq, err := cfg.client.GetConsentRequest(challenge)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var redirectTo string
	if consentReq.Skip {
		log.Info("Accepting consent request")
		completedReq, err := cfg.client.AcceptConsentRequest(challenge, &models.AcceptConsentRequest{
			GrantScope:               consentReq.RequestedScope,
			GrantAccessTokenAudience: consentReq.RequestedAccessTokenAudience,
			Session: &models.ConsentRequestSession{
				// ???
			},
		})

		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		redirectTo = *completedReq.RedirectTo
		http.Redirect(w, req, redirectTo, http.StatusFound)

	} else {
		requestedScope := consentReq.RequestedScope
		log.Println(requestedScope)
		//display consent page here - grant permissions
	}
}
