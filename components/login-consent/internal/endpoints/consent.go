package endpoints

import (
	"net/http"

	"github.com/kyma-project/kyma/components/login-consent/internal/helpers"
	"github.com/ory/hydra-client-go/models"
	log "github.com/sirupsen/logrus"
)

type loginContext struct {
	accessToken string
	idToken     string
}

func (cfg *Config) Consent(w http.ResponseWriter, req *http.Request) {

	log.Infof("DEBUG: Consent endpoint hit with req.URL: %s", req.URL.String())
	challenge, err := helpers.GetLConsentChallenge(req)

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
	consentReq, err := cfg.hydraClient.GetConsentRequest(challenge)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Infof("consentReqContext: %s", consentReq.Context)

	//var loginContext loginContext

	var redirectTo string
	//if consentReq.Skip {
	log.Info("Accepting consent request")
	completedReq, err := cfg.hydraClient.AcceptConsentRequest(challenge, &models.AcceptConsentRequest{
		GrantScope:               consentReq.RequestedScope,
		GrantAccessTokenAudience: consentReq.RequestedAccessTokenAudience,
		Session: &models.ConsentRequestSession{
			IDToken: consentReq.Context,
		},
	})

	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	redirectTo = *completedReq.RedirectTo
	log.Infof("DEBUG: redirecting to: %s", redirectTo)
	http.Redirect(w, req, redirectTo, http.StatusFound)

	//} else {
	//requestedScope := consentReq.RequestedScope
	//log.Println(requestedScope)
	//display consent page here - grant permissions
	//}
}
