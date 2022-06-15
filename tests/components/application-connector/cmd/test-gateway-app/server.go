package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"io/ioutil"
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

	basicAuthCredentials := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}

	test_api.AddAPIHandler(apiRouter, oauthTokesCache, csrfTokensCache, mutex, basicAuthCredentials)

	oauthCredentials := test_api.OAuthCredentials{ClientID: cfg.OAuthClientID, ClientSecret: cfg.OAuthClientSecret}
	test_api.AddOAuthTokensHandler(apiRouter, oauthTokesCache, csrfTokensCache, oauthCredentials, mutex)

	tokensRouter := mux.NewRouter()
	test_api.AddOAuthTokensHandler(tokensRouter, oauthTokesCache, csrfTokensCache, oauthCredentials, mutex)

	// TODO Use https://github.com/oklog/run instead.
	// Note: it is used here: https://github.com/kyma-project/kyma/blob/main/components/central-application-connectivity-validator/cmd/centralapplicationconnectivityvalidator/centralapplicationconnectivityvalidator.go
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		address := fmt.Sprintf(":%d", cfg.MtlsPort)
		mtlsServer := newMTLSServer(cfg.CaCertPath, address, tokensRouter)
		log.Fatal(mtlsServer.ListenAndServeTLS(cfg.ServerCertPath, cfg.ServerKeyPath))
		wg.Done()
	}()

	go func() {
		address := fmt.Sprintf(":%d", cfg.Port)
		log.Fatal(http.ListenAndServe(address, apiRouter))
		wg.Done()
	}()

	wg.Wait()
}

func newMTLSServer(caCertPath, address string, handler http.Handler) *http.Server {
	// Create a CA certificate pool and add cert.pem to it
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create the TLS Config with the CA pool and enable Client certificate validation
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return &http.Server{
		Addr:      address,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
