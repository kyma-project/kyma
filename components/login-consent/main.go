package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/kyma/components/login-consent/internal/endpoints"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	var appPort string
	var appAddress string
	var hydraAddr string
	var hydraPort string
	var dexAddress string
	var clientID string
	var clientSecret string

	flag.StringVar(&appPort, "app-port", "3000","Application port")
	flag.StringVar(&appAddress, "app-address", "https://ory-login-consent.jk6.goatz.shoot.canary.k8s-hana.ondemand.com","Application address")
	flag.StringVar(&hydraAddr, "hydra-address", "http://ory-hydra-admin.kyma-system.svc.cluster.local", "Hydra administrative endpoint address")
	flag.StringVar(&hydraPort, "hydra-port", "4445", "Hydra administrative endpoint port")
	flag.StringVar(&dexAddress, "dex-address", "https://dex.jk6.goatz.shoot.canary.k8s-hana.ondemand.com", "Dex address")
	flag.StringVar(&clientID, "client-id", "go-consent-app", "Client ID")
	flag.StringVar(&clientSecret, "client-secret", "go-consent-secret", "Client secret")

	flag.Parse()

	redirectURL := fmt.Sprintf("%s/callback", appAddress)
	scopes := []string{"email", "openid", "profile", "groups"}
	authn, err := endpoints.NewAuthenticator(dexAddress, clientID, clientSecret, redirectURL, scopes)
	if err != nil {
		log.Error(err)
	}

	cfg, err := endpoints.New(hydraAddr, hydraPort, authn)
	if err != nil {
		log.Error(err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/login", cfg.Login)
	m.HandleFunc("/callback", cfg.Callback)
	m.HandleFunc("/consent", cfg.Consent)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", appPort),
		Handler: m,
	}

	log.Infof("Starting server on port %s", appPort)
	err = srv.ListenAndServe()
	if err != nil {
		log.Info(err)
	}
}
