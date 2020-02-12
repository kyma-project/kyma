package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/eventing/pkg/client/informers/externalversions"
	"knative.dev/eventing/pkg/client/listers/eventing/v1alpha1"

	"github.com/kyma-project/kyma/components/event-service/internal/events/mesh"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	"github.com/kyma-project/kyma/components/event-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/signals"
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

	if err := mesh.Init(options.sourceID, options.eventMeshURL); err != nil {
		log.Errorf("failed to initialize the Event mesh configuration")
		os.Exit(1)
	}

	knClient, e := getKnativeClient()
	if e != nil {
		log.Error("unable to get Knative client", e.Error())
		return
	}

	eventsClient := subscribed.NewEventsClient(knClient)

	externalHandler := externalapi.NewHandler(options.maxRequestSize, eventsClient, options.eventMeshURL)

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

func getKnativeClient() (versioned.Interface, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	knEventingClient, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		log.Infof("error in creating knative client: %+v", err)
		return nil, err
	}

	return knEventingClient, nil
}

// TODO(marcobebway) remove this
func initKnativeTriggerLister() (v1alpha1.TriggerLister, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	knEventingClient, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		log.Infof("error in creating knative client: %+v", err)
		return nil, err
	}

	informerFactory := externalversions.NewSharedInformerFactory(knEventingClient, 0)
	ctx := signals.NewContext()
	go informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	return informerFactory.Eventing().V1alpha1().Triggers().Lister(), nil
}
