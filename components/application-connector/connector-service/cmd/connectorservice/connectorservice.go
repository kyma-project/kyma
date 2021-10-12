package main

import (
	"net/http"
	"strconv"
	"sync"

	loggingMiddlewares "github.com/kyma-project/kyma/components/connector-service/internal/logging/middlewares"
	"github.com/kyma-project/kyma/components/connector-service/internal/monitoring"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens"
	"github.com/kyma-project/kyma/components/connector-service/internal/tokens/tokencache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Certificate Service.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	env := parseEnv()
	log.Infof("Environment variables: %s", env)

	tokenCache := tokencache.NewTokenCache()
	tokenGenerator := tokens.NewTokenGenerator(options.tokenLength)
	tokenManager := tokens.NewTokenManager(tokenCache)
	tokenCreatorProvider := tokens.NewTokenCreatorProvider(tokenCache, tokenGenerator.NewToken)

	globalMiddlewares, appErr := monitoring.SetupMonitoringMiddleware()
	if appErr != nil {
		log.Errorf("Error while setting up monitoring: %s", appErr)
	}

	if options.requestLogging {
		globalMiddlewares = append(globalMiddlewares, loggingMiddlewares.NewRequestLoggingMiddleware().Middleware)
	}

	handlers := createAPIHandlers(tokenManager, tokenCreatorProvider, options, env, globalMiddlewares)

	externalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.externalAPIPort),
		Handler: handlers.externalAPI,
	}

	internalSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(options.internalAPIPort),
		Handler: handlers.internalAPI,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	go func() {
		log.Info(internalSrv.ListenAndServe())
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info(http.ListenAndServe(":9090", nil))
	}()

	wg.Wait()
}
