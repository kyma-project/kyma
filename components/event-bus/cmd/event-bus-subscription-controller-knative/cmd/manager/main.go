package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-logr/logr"
	pushv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	eav1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	subscriptionController "github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	sckOpts := opts.ParseFlags()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")

	// Get a config to talk to the apiserver
	log.Info("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	log.Info("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("setting up scheme")
	if err := pushv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add push APIs to scheme")
		os.Exit(1)
	}
	if err := eav1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add event activation APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	log.Info("Setting up subscription controller")
	if err := subscriptionController.AddToManager(mgr, sckOpts); err != nil {
		log.Error(err, "unable to register subscription controller to the manager")
		os.Exit(1)
	}

	_, err = eventactivation.ProvideController(mgr)
	if err != nil {
		log.Error(err,"Unable to create Event Activation controller")
		os.Exit(1)
	}

	// Set up healthcheck handlers
	serveMux := http.NewServeMux()
	serveMux.Handle("/v1/status/live", statusLiveHandler(&mgr, log))
	serveMux.Handle("/v1/status/ready", statusReadyHandler(&mgr, log))

	// Start HTTP server for healthchecks
	log.Info("HTTP server starting on", "port", sckOpts.Port)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", sckOpts.Port), serveMux); err != nil {
			log.Error(err, "HTTP server failed with error")
		}
	}()

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt,
		syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			println("Subscription controller manager signaled with: os.Interrupt")
			println("Signal: ", sig)
		case syscall.SIGTERM:
			println("Subscription controller manager signaled with: SIGTERM")
			println("Signal: ", sig)
		case syscall.SIGHUP:
			println("Subscription controller manager signaled with: SIGHUP")
			println("Signal: ", sig)
		case syscall.SIGINT:
			println("Subscription controller manager signaled with: SIGINT")
			println("Signal: ", sig)
		case syscall.SIGQUIT:
			println("Subscription controller manager signaled with: SIGQUIT")
			println("Signal: ", sig)
		}
	}()

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}

}

var statusLive, statusReady common.StatusReady

func statusLiveHandler(manager *manager.Manager, log logr.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager != nil {
			if statusLive.SetReady() {
				log.Info("statusLiveHandler :: Status: READY")
			}
			w.WriteHeader(http.StatusOK)
		} else {
			statusLive.SetNotReady()
			log.Info("statusLiveHandler :: Status: NOT_READY")
			w.WriteHeader(http.StatusBadGateway)
		}
	})
}

func statusReadyHandler(manager *manager.Manager, log logr.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if manager != nil {
			if statusReady.SetReady() {
				log.Info("statusReadyHandler :: Status: READY")
			}
			w.WriteHeader(http.StatusOK)
		} else {
			statusReady.SetNotReady()
			log.Info("statusReadyHandler :: Status: NOT_READY")
			w.WriteHeader(http.StatusBadGateway)
		}
	})
}
