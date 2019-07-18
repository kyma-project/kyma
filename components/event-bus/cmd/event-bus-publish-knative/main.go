package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/application"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/httpserver"
	"github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish-knative/publisher"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/publish/opts"
	knative "github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	"github.com/kyma-project/kyma/components/event-bus/internal/trace"
)

const (
	shutdownTimeout = time.Minute
)

func main() {
	// read application options from the cli args
	options := opts.ParseFlags()

	// init the knative lib
	knativeLib, err := knative.GetKnativeLib()
	if err != nil {
		log.Fatalf("failed to start knative publish application: %v", err)
	}

	// init the knative publisher
	knativePublisher := publisher.NewKnativePublisher()

	// init the tracer
	tracer := trace.StartNewTracer(options.TraceOptions)
	defer tracer.Stop()

	// start the Knative publish application
	app := application.StartKnativePublishApplication(options, knativeLib, &knativePublisher, &tracer)
	log.Println("started knative publish application")

	// print the application startup options
	options.Print()

	// init the HTTP max concurrent requests
	handler, semaphore := limit(app.ServeMux(), &options.MaxRequests)
	defer close(semaphore)

	// start the HTTP server
	server := httpserver.NewHTTPServer(&options.Port, &handler)
	go server.Start()

	metricsServeMux := http.NewServeMux()
	metricsServeMux.Handle("/metrics", promhttp.Handler())
	metricsServer := http.Server{
		Addr:    fmt.Sprintf(":%d", options.MonitoringPort),
		Handler: metricsServeMux,
	}
	log.Printf("HTTP metrics server starting on %v", options.MonitoringPort)
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Fatalf("HTTP metrics server failed with error:\n%v", err)
		}
	}()

	// shutdown the HTTP server gracefully
	server.Shutdown(shutdownTimeout)
}

func limit(serveMux *http.ServeMux, maxRequests *int) (http.Handler, chan bool) {
	// create a channel to limit the max requests handled concurrently
	semaphore := make(chan bool, *maxRequests)

	// init the handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		semaphore <- true
		defer func() { <-semaphore }()
		serveMux.ServeHTTP(w, r)
	})

	return handler, semaphore
}
