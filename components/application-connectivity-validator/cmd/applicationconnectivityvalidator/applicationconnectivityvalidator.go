package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/common/logging/tracing"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/externalapi"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/validationproxy"
	"github.com/patrickmn/go-cache"
)

func main() {
	options, err := parseOptions()
	if err != nil {
		if logErr := logger.LogFatalError("Failed to parse options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s,Failed to parse options: %s", logErr, err)
		}
		os.Exit(1)
	}

	level, err := logger.MapLevel(options.LogLevel)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to map log level from options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to map log level from options: %s", logErr, err)
		}

		os.Exit(2)
	}
	format, err := logger.MapFormat(options.LogFormat)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to map log format from options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to map log format from options: %s", logErr, err)
		}
		os.Exit(3)
	}
	log, err := logger.New(format, level)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to initialize logger: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to initialize logger: %s", logErr, err)
		}
		os.Exit(4)
	}
	if err := logger.InitKlog(log, level); err != nil {
		log.WithContext().Error("While initializing klog logger: %s", err.Error())
		os.Exit(5)
	}

	log.WithContext().With("options", options).Info("Starting Validation Proxy.")

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

	tracingMiddleware := tracing.NewTracingMiddleware(proxyHandler.ProxyAppConnectorRequests)

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
		controller.Start(log, options.kubeConfig, options.mainURL, options.syncPeriod, options.appName, idCache)
	}()

	go func() {
		log.WithContext().With("server", "proxy").With("port", options.proxyPort).Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.WithContext().With("server", "external").With("port", options.externalAPIPort).Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}
