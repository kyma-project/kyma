package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"knative.dev/pkg/injection/sharedmain"

	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/eventactivation"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/knativesubscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/controller/subscription"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/subscription/opts"
)

func main() {
	sckOpts := opts.ParseFlags()

	log.SetLogger(log.ZapLogger(false))
	log := log.Log.WithName("entrypoint")

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

	sharedmain.Main("eventbus_controller",
		eventactivation.NewController,
		subscription.NewController,
		knativesubscription.NewController)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

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
