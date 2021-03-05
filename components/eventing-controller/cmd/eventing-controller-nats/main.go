package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	subscription "github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription-nats"
)

func main() {
	setupLog := ctrl.Log.WithName("setup")

	var metricsAddr string
	var probeAddr string
	var readyEndpoint string
	var healthEndpoint string
	var enableDebugLogs bool
	var maxReconnects int
	var reconnectWait time.Duration
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&readyEndpoint, "ready-check-endpoint", "readyz", "The endpoint of the readiness probe.")
	flag.StringVar(&healthEndpoint, "health-check-endpoint", "healthz", "The endpoint of the health probe.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.IntVar(&maxReconnects, "max-reconnects", 10, "Maximum number of reconnect attempts.")
	flag.DurationVar(&reconnectWait, "reconnect-wait", time.Second, "Wait time between reconnect attempts.")
	flag.Parse()

	cfg := env.GetNatsConfig(maxReconnects, reconnectWait)
	if len(cfg.Url) == 0 {
		setupLog.Error(fmt.Errorf("env var URL should be a non-empty value"), "unable to start manager")
		os.Exit(1)
	}
	setupLog.Info("Nats config", "URL", cfg.Url)

	scheme, err := setupScheme()
	if err != nil {
		setupLog.Error(err, "failed to setup scheme")
		os.Exit(1)
	}

	// cluster config
	k8sConfig := ctrl.GetConfigOrDie()

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDebugLogs)))
	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		HealthProbeBindAddress: probeAddr,
		Port:                   9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// setup application lister
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	if err := subscription.NewReconciler(
		mgr.GetClient(),
		applicationLister,
		mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		mgr.GetEventRecorderFor("eventing-controller-nats"),
		cfg,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup the NATS Subscription controller")
		cancel()
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck(healthEndpoint, healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck(readyEndpoint, healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start manager")
		cancel()
		os.Exit(1)
	}
}

func setupScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}
