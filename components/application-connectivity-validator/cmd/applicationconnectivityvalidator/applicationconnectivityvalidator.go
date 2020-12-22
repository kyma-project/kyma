package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/validationproxy"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Validation Proxy.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	idCache := cache.New(
		time.Duration(options.cacheExpirationMinutes)*time.Minute,
		time.Duration(options.cacheCleanupMinutes)*time.Minute,
	)

	proxyHandler := validationproxy.NewProxyHandler(
		options.group,
		options.tenant,
		options.eventServicePathPrefixV1,
		options.eventServicePathPrefixV2,
		options.eventServiceHost,
		options.eventMeshPathPrefix,
		options.eventMeshHost,
		options.eventMeshDestinationPath,
		options.appRegistryPathPrefix,
		options.appRegistryHost,
		idCache)

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
		// TODO: go routine should inform other go routines that it initially updated the cache
		controller.Start(options.kubeConfig, options.masterURL, options.syncPeriod, options.appName, idCache)
	}()

	go func() {
		log.Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}
