package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
)

func main() {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load Authorization server config")
	//TODO: Init it correctly
	cfg.RequestHeaders = map[string][]string{"hkey1":{"hval1"},"hkey2":{"hval21","hval22"}}
	cfg.RequestQueryParameters = map[string][]string{"qkey1":{"qval1"},"qkey2":{"qval21","qval22"}}

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level: '%s', defaulting to 'info'", cfg.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Infof("Starting frog auth server application")
	log.Infof("Config: %s", cfg.String())

	basicAuthCredentials := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}
	oAuthCredentials := test_api.OAuthCredentials{ClientID: cfg.OAuthClientID, ClientSecret: cfg.OAuthClientSecret}
	expectedRequestParameters := test_api.ExpectedRequestParameters{Headers: cfg.RequestHeaders, QueryParameters: cfg.RequestQueryParameters}

	router := test_api.SetupRoutes(os.Stdout, basicAuthCredentials, oAuthCredentials, expectedRequestParameters)

	address := fmt.Sprintf(":%d", cfg.Port)
	log.Fatal(http.ListenAndServe(address, router))
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
