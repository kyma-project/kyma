package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/kyma-project/kyma/components/login-consent/internal/endpoints"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	log "github.com/sirupsen/logrus"
)

func main() {
	var appPort string
	var appAddress string
	var hydraAddr string
	var hydraPort string
	var dexAddress string
	var clientID string
	var clientSecret string

	flag.StringVar(&appPort, "app-port", "3000", "Application port")
	flag.StringVar(&appAddress, "app-address", "https://ory-login-consent.ts-1.goatz.shoot.canary.k8s-hana.ondemand.com", "Application address")
	flag.StringVar(&hydraAddr, "hydra-address", "http://ory-hydra-admin.kyma-system.svc.cluster.local", "Hydra administrative endpoint address")
	flag.StringVar(&hydraPort, "hydra-port", "4445", "Hydra administrative endpoint port")
	flag.StringVar(&dexAddress, "dex-address", "https://dex.ts-1.goatz.shoot.canary.k8s-hana.ondemand.com", "Dex address")
	flag.StringVar(&clientID, "client-id", "hydra-consent-app", "client ID")
	flag.StringVar(&clientSecret, "client-secret", "example-app-secret", "client secret")

	flag.Parse()

	log.Info("appPort: ", appPort)
	log.Info("appAddress: ", appAddress)
	log.Info("hydraAddr: ", hydraAddr)
	log.Info("hydraPort: ", hydraPort)
	log.Info("dexAddress: ", dexAddress)
	log.Info("clientID: ", clientID)
	log.Info("clientSecret: ", clientSecret)

	//Setup Authenticator
	redirectURL := fmt.Sprintf("%s/callback", appAddress)
	scopes := []string{"email", "openid", "profile", "groups"}
	authn, err := endpoints.NewAuthenticator(dexAddress, clientID, clientSecret, redirectURL, scopes)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	//Setup Hydra Client
	rawHydraURL := fmt.Sprintf("%s:%s/", hydraAddr, hydraPort)
	hydraURL, err := url.Parse(rawHydraURL)
	if err != nil {
		log.Errorf("failed to parse Hydra url: %s", err)
		os.Exit(1)
	}
	hydraClient := hydra.NewClient(&http.Client{}, *hydraURL, "https")

	cfg, err := endpoints.New(&hydraClient, authn)
	if err != nil {
		log.Error(err)
		os.Exit(1)
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
