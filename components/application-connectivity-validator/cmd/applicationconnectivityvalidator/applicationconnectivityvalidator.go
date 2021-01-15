package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/tracing"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/validationproxy"
	logger "github.com/kyma-project/kyma/components/application-connectivity-validator/pkg/logger"
	"github.com/patrickmn/go-cache"
)

func main() {
	options := parseArgs()
	//TODO: what to do with this
	fmt.Printf("Options: %+v:\n", options)
	level, err := logger.MapLevel(options.logLevel)
	if err != nil {
		logger.LogFatalError("Fatal Error: %s", err.Error())
		os.Exit(1)
	}
	format, err := logger.MapFormat(options.logFormat)
	if err != nil {
		logger.LogFatalError("Fatal Error: %s", err.Error())
		os.Exit(1)
	}

	level = logger.INFO
	log := logger.New(format, level)
	log.Info("Starting Validation Proxy.")

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
		idCache,
		log)

	tracingMiddleware := tracing.NewTracingMiddleware(log, proxyHandler.ProxyAppConnectorRequests)

	proxyServer := http.Server{
		Handler: validationproxy.NewHandler(tracingMiddleware),
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
		controller.Start(log, options.kubeConfig, options.masterURL, options.syncPeriod, options.appName, idCache)
	}()

	go func() {
		log.Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}
