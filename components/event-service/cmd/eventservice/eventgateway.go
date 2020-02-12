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

	subscriptions "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-service/internal/events/mesh"
	"github.com/kyma-project/kyma/components/event-service/internal/events/subscribed"
	"github.com/kyma-project/kyma/components/event-service/internal/externalapi"
	"github.com/kyma-project/kyma/components/event-service/internal/httptools"
	log "github.com/sirupsen/logrus"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
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

	subscriptionsClient, namespacesClient, triggerLister, e := initK8sResourcesClients()

	if e != nil {
		log.Error("Unable to create Events Client.", e.Error())
		return
	}

	eventsClient := subscribed.NewEventsClient(subscriptionsClient, namespacesClient, triggerLister)

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

func initK8sResourcesClients() (subscribed.SubscriptionsGetter, subscribed.NamespacesClient, v1alpha1.TriggerLister, error) {
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	subscriptionsClient, err := subscriptions.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	//TODO(marcobebway) Check if we can clear the namespace client
	coreClient, err := core.NewForConfig(k8sConfig)
	if err != nil {
		return nil, nil, nil, err
	}

	namespacesClient := coreClient.Namespaces()

	knEventingClient, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		log.Infof("error in creating knative client: %+v", err)
		return nil, nil, nil, err
	}

	informerFactory := externalversions.NewSharedInformerFactory(knEventingClient, 0)
	ctx := signals.NewContext()
	go informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	return subscriptionsClient.EventingV1alpha1(), namespacesClient, informerFactory.Eventing().V1alpha1().Triggers().Lister(), nil
}
