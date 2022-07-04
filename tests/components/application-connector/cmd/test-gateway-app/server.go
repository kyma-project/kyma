package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
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

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level: '%s', defaulting to 'info'", cfg.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Infof("Starting frog auth server application")
	log.Infof("Config: %s", cfg.String())

	basicAuthCredentials := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}

	router := test_api.SetupRoutes(os.Stdout, basicAuthCredentials)

	address := fmt.Sprintf(":%d", cfg.Port)
	log.Fatal(http.ListenAndServe(address, router))
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
