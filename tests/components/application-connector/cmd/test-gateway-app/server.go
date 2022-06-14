package main

import (
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"net/http"
	"sync"
)

func main() {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load Authorization server config")

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level: '%s', defaulting to 'info'", cfg.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Infof("Starting frog auth server application")
	log.Infof("Config: %s", cfg.String())

	oauthTokesCache := make(map[string]bool)
	csrfTokensCache := make(map[string]bool)
	mutex := &sync.RWMutex{}

	apiRouter := mux.NewRouter()

	basciAuthCredentails := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}

	test_api.AddAPIHandler(apiRouter, oauthTokesCache, csrfTokensCache, mutex, basciAuthCredentails)

	oauthCredentials := test_api.OAuthCredentials{ClientID: cfg.OAuthClientID, ClientSecret: cfg.OAuthClientSecret}
	test_api.AddTokensHandler(apiRouter, oauthTokesCache, csrfTokensCache, oauthCredentials, mutex)

	tokensRouter := mux.NewRouter()
	test_api.AddTokensHandler(tokensRouter, oauthTokesCache, csrfTokensCache, oauthCredentials, mutex)

	// TODO This implementation must be fixed

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		log.Fatal(http.ListenAndServeTLS(":"+string(cfg.MtlsPort), "/etc/secret-volume/tls.crt", "./etc/secret-volume/tls.key", tokensRouter))
		wg.Done()
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":"+string(cfg.Port), apiRouter))
		wg.Done()
	}()

	wg.Wait()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
