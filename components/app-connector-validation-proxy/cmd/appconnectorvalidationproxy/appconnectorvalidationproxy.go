package main

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/app-connector-validation-proxy/internal/validationproxy"
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

	server := http.Server{
		Handler: validationproxy.NewHandler(proxyHandler),
		Addr:    fmt.Sprintf(":%d", options.proxyPort),
	}

	log.Error(server.ListenAndServe())
}
