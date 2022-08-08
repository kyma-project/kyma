package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/test-api"
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

	log.Infof("Starting mock application")
	log.Infof("Config: %s", cfg.String())

	wg := sync.WaitGroup{}
	wg.Add(3)

	basicAuthCredentials := test_api.BasicAuthCredentials{User: cfg.BasicAuthUser, Password: cfg.BasicAuthPassword}
	oAuthCredentials := test_api.OAuthCredentials{ClientID: cfg.OAuthClientID, ClientSecret: cfg.OAuthClientSecret}
	expectedRequestParameters := test_api.ExpectedRequestParameters{Headers: cfg.RequestHeaders, QueryParameters: cfg.RequestQueryParameters}
	oauthTokens := make(map[string]test_api.OAuthToken)
	csrfTokens := make(map[string]interface{})

	go func() {
		address := fmt.Sprintf(":%d", cfg.Port)
		router := test_api.SetupRoutes(os.Stdout, basicAuthCredentials, oAuthCredentials, expectedRequestParameters, oauthTokens, csrfTokens)
		log.Fatal(http.ListenAndServe(address, router))
	}()

	go func() {
		address := fmt.Sprintf(":%d", cfg.mTLS.port)
		router := test_api.SetupMTLSRoutes(os.Stdout, oAuthCredentials, oauthTokens, csrfTokens)
		mtlsServer := newMTLSServer(cfg.mTLS.caCertPath, address, router)
		log.Fatal(mtlsServer.ListenAndServeTLS(cfg.mTLS.serverCertPath, cfg.mTLS.serverKeyPath))
		wg.Done()
	}()

	go func() {
		address := fmt.Sprintf(":%d", cfg.mTLSExpiredCerts.port)
		router := test_api.SetupMTLSRoutes(os.Stdout, oAuthCredentials, oauthTokens, csrfTokens)
		mtlsServer := newMTLSServer(cfg.mTLSExpiredCerts.caCertPath, address, router)
		log.Fatal(mtlsServer.ListenAndServeTLS(cfg.mTLSExpiredCerts.serverCertPath, cfg.mTLSExpiredCerts.serverKeyPath))
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
