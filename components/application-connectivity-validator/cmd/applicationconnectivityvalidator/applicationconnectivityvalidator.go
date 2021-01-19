package main

import (
	"context"
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
	logger.LogOptions("Parsed options: %+v", options)

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
	log := logger.New(format, level)

	logger.InitKlog(log, level)
	log.WithContext().Info("Starting Validation Proxy.")
	log.WithContext().Error("this is sample error log")

	ctx := context.TODO()
	ctx = context.WithValue(ctx, "traceid", "abc")
	ctx = context.WithValue(ctx, "spanid", "def")

	// Body handlera, gdzie przekazujemy log

	log.WithTracing(ctx).With("key", "val").With("key2", "val2").Errorf("some error1")

	// Body controllera, przekazujemy log.WithContext()
	log.WithContext().With("key1", "val2").Errorf("some error2")
	log.WithContext().Errorf("some error2")

	//log.WithFields(tracing.GetMetadata(ctx)).Error("sample log")
	//log.WithFields(tracing.GetMetadata(ctx)).Error("sample second log")
	//log.WithFields(tracing.GetMetadata(ctx)).
	//	WithFields(tracing.GetMetadata(ctx)).
	//	WithContext(map[string]string{"key": "val"}).
	//	EnhanceContext(map[string]string{"key2": "val2"}).
	//	EnhanceContext(map[string]string{"key3": "val3"}). //można dodawać kilka razy
	//	EnhanceContext(map[string]string{"key2": "val8"}). //nie podmienia, tylko dodaje
	//	Error("Two Tracings and Two Contexts")
	//log.EnhanceContext(map[string]string{"key": "val"})
	//log.EnhanceContext(map[string]string{"key2": "val2"}).Infof("Only EnhanceContext")
	//
	//log.WithContext(map[string]string{"key3": "val3"})
	//log.EnhanceContext(map[string]string{"key4": "val4"}).Infof("Second chance")
	//
	//log.WithContext(map[string]string{"key3": "val3"}).
	//	EnhanceContext(map[string]string{"key4": "val4"}).
	//	With("dupa1", "dupa2").
	//	Infof("Third chance")

	//log.WithFields(tracing.GetMetadata(ctx)).WithFields(tracing.GetMetadata(ctx)).Error("overwrite context")

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
		controller.Start(log.WithContext(), options.kubeConfig, options.masterURL, options.syncPeriod, options.appName, idCache)
	}()

	go func() {
		log.WithContext().Error(proxyServer.ListenAndServe())
	}()

	go func() {
		log.WithContext().Error(externalServer.ListenAndServe())
	}()

	wg.Wait()
}
