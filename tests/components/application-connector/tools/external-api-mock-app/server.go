package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
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

	wg := sync.WaitGroup{}
	wg.Add(2)

	basicAuthCredentials := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}
	oAuthCredentials := test_api.OAuthCredentials{ClientID: cfg.OAuthClientID, ClientSecret: cfg.OAuthClientSecret}

	go func() {
		expectedRequestParameters := test_api.ExpectedRequestParameters{Headers: cfg.RequestHeaders, QueryParameters: cfg.RequestQueryParameters}

		router := test_api.SetupRoutes(os.Stdout, basicAuthCredentials, oAuthCredentials, expectedRequestParameters)

		address := fmt.Sprintf(":%d", cfg.Port)
		log.Fatal(http.ListenAndServe(address, router))
	}()

	go func() {
		address := fmt.Sprintf(":%d", cfg.MtlsPort)
		router := test_api.SetupMTLSRoutes(os.Stdout, oAuthCredentials)
		mtlsServer := newMTLSServer(cfg.CaCertPath, address, router)
		log.Fatal(mtlsServer.ListenAndServeTLS(cfg.ServerCertPath, cfg.ServerKeyPath))
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
		ClientCAs:          caCertPool,
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true,
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
