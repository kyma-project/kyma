package main

import (
	"net/http"
	"strconv"
	"time"

	"sync"

	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/httptools"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-connector/application-registry/internal/monitoring"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting metadata.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	nameResolver := k8sconsts.NewNameResolver(options.namespace, options.centralGatewayUrl)

	serviceDefinitionService, err := newServiceDefinitionService(
		options,
		nameResolver,
	)

	if err != nil {
		log.Errorf("Unable to initialize Metadata Service, %s", err.Error())
	}

	middlewares, err := monitoring.SetupMonitoringMiddleware()
	if err != nil {
		log.Errorf("Failed to setup monitoring middleware, %s", err.Error())
	}

	externalHandler := newExternalHandler(serviceDefinitionService, middlewares, options)

	if options.requestLogging {
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info(http.ListenAndServe(":9090", nil))
	}()

	wg.Wait()
}
