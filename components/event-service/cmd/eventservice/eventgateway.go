package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kyma-project/kyma/components/event-service/internal/events/bus"
	"github.com/kyma-project/kyma/components/event-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
	log "github.com/sirupsen/logrus"
)

const (
	shutdownTimeout = time.Minute
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting event-service")

	options := parseArgs()
	log.Infof("Options: %s", options)

	bus.Init(options.sourceID, options.eventsTargetURL)

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

	// start the HTTP server
	go start(externalSrv)

	// shutdown the HTTP server gracefully
	shutdown(externalSrv, shutdownTimeout)
}

func start(server *http.Server) {
	if server == nil {
		log.Error("cannot start a nil HTTP server")
		return
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error(err)
	}
}

func shutdown(server *http.Server, timeout time.Duration) {
	if server == nil {
		log.Info("cannot shutdown a nil HTTP server")
		return
	}

	shutdownSignal := make(chan os.Signal, 1)
	defer close(shutdownSignal)

	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-shutdownSignal

	log.Infof("HTTP server shutdown with timeout: %s\n", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server shutdown error: %v\n", err)
	} else {
		log.Info("HTTP server shutdown successful")
	}
}
