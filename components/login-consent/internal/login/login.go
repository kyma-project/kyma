package login

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	hydraAPI "github.com/ory/hydra-client-go/models"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
)

type HydraConfig struct {
	hydraAddr string
	hydraPort string
}

func NewHydraConfig(hydraAddr string, hydraPort string) *HydraConfig {
	return &HydraConfig{
		hydraAddr: hydraAddr,
		hydraPort: hydraPort,
	}
}

func (hc *HydraConfig) Callback(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Callback hit"))
	w.WriteHeader(http.StatusOK)
}

func (hc *HydraConfig) Login(w http.ResponseWriter, req *http.Request) {
	challenge, err := getLoginChallenge(req.URL.Query())
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	loginReq := hydra.GetLoginRequest(challenge)

	var redirectTo string
	if *loginReq.Skip == true {
		response := hydra.AcceptLoginRequest(challenge, hydraAPI.AcceptLoginRequest{
			Remember:    true,
			RememberFor: 30,
			Subject:     nil,
		})
		redirectTo = *response.RedirectTo
	} else {
		ctx := context.Background()
		issuer, err := oidc.NewProvider(ctx, "https://dex.jk6.goatz.shoot.canary.k8s-hana.ondemand.com") //TODO: provide dex addr here
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		client := oauth2.Config{
			ClientID:     "go-consent-app", //TODO: provide proper data here
			ClientSecret: "go-consent-secret",
			Endpoint:     issuer.Endpoint(),
			RedirectURL:  "http://localhost:8080/callback",
			Scopes:       []string{"openid", "email", "profile", "groups"},
		}
		redirectTo = client.AuthCodeURL("state")
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
