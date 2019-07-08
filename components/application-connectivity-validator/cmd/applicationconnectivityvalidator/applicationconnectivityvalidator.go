package main

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/externalapi"
	"net/http"
	"sync"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/validationproxy"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Validation Proxy.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	proxyHandler := validationproxy.NewProxyHandler(
		options.group,
		options.tenant,
		options.eventServicePathPrefix,
		options.eventServiceHost,
		options.appRegistryPathPrefix,
		options.appRegistryHost)

	proxyServer := http.Server{
		Handler: validationproxy.NewHandler(proxyHandler),
		Addr:    fmt.Sprintf(":%d", options.proxyPort),
	}

	externalServer := http.Server{
		Handler: externalapi.NewHandler(),
		Addr:    fmt.Sprintf(":%d", options.externalAPIPort),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}
