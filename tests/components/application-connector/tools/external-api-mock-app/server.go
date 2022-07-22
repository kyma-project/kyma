package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := NewConfig()
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
