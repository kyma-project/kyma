package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
	log "github.com/sirupsen/logrus"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting event-service")

	options := parseArgs()
	log.Infof("Options: %s", options)

	bus.Init(options.sourceNamespace, options.sourceType, options.sourceEnvironment, options.eventsTargetURL)

	externalHandler := externalapi.NewHandler()

	if options.requestLogging {
		externalHandler = httptools.RequestLogger("External handler: ", externalHandler)
	}

	externalSrv := &http.Server{
		Addr:         ":" + strconv.Itoa(options.externalAPIPort),
		Handler:      externalHandler,
		ReadTimeout:  time.Duration(options.requestTimeout) * time.Second,
		WriteTimeout: time.Duration(options.requestTimeout) * time.Second,
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Info(externalSrv.ListenAndServe())
	}()

	wg.Wait()
}
