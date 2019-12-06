package main

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	evapisv1alpha1 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	pushv1alpha1 "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	eav1alpha1 "github.com/kyma-project/kyma/components/event-bus/internal/ea/apis/applicationconnector.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/knativesubscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"

	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	if err := messagingV1Alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add event activation APIs to scheme")
		os.Exit(1)
	}
	if err := evapisv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add Knative Eventing APIs to scheme")
		os.Exit(1)
	}
	// Setup all Controllers
	log.Info("Setting up Subscription Controller")
	err = subscription.ProvideController(mgr, sckOpts)
	if err != nil {
		log.Error(err, "unable to create Subscription controller")
		os.Exit(1)
	}

	log.Info("Setting up Event Activation Controller")
	err = eventactivation.ProvideController(mgr)
	if err != nil {
		log.Error(err, "unable to create Event Activation controller")
		os.Exit(1)
	}

	log.Info("Setting up Knative Subscription Controller")
	err = knativesubscription.ProvideController(mgr)
	if err != nil {
		log.Error(err, "unable to create Knative Subscription controller")
		os.Exit(1)
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// Start the Manager
	log.Info("Starting the manager.")
	go func() {
		if err := mgr.Start(stopCh); err != nil {
			log.Error(err, "unable to run the manager")
			os.Exit(1)
		}
	}()

	// Setup Monitoring Service
	metricsServeMux := http.NewServeMux()
	metricsServeMux.Handle("/metrics", promhttp.Handler())
	metricsServer := http.Server{
		Addr:    fmt.Sprintf(":%d", 9090),
		Handler: metricsServeMux,
	}
	log.Info("HTTP metrics server starting on %v", 9090)
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil {
			log.Error(err, "HTTP metrics server failed with error|||")
		}
	}()
	// Setup health check handlers
	serveMux := http.NewServeMux()
	serveMux.Handle("/v1/status/live", statusLiveHandler(&mgr, log))
	serveMux.Handle("/v1/status/ready", statusReadyHandler(&mgr, log))
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", sckOpts.Port),
		Handler: serveMux,
	}

	// Start HTTP server for health checks
	log.Info("HTTP server starting on", "port", sckOpts.Port)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Error(err, "HTTP server failed with error")
			os.Exit(1)
		}
	}()

	<-stopCh

	log.Info("Signal received, shutting down gracefully")
	// Close the http server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
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
