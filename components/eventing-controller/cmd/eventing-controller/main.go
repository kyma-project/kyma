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

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = eventingv1alpha1.AddToScheme(scheme)
	_ = apigatewayv1alpha1.AddToScheme(scheme)
}

func main() {
	setupLog := ctrl.Log.WithName("setup")

	var metricsAddr string
	var probeAddr string
	var readyEndpoint string
	var healthEndpoint string
	var resyncPeriod time.Duration
	var enableDebugLogs bool
	flag.StringVar(&metricsAddr, "metrics-address", ":8080", "The TCP address that the controller should bind to for serving prometheus metrics.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The TCP address that the controller should bind to for serving health probes.")
	flag.StringVar(&readyEndpoint, "ready-check-endpoint", "readyz", "The endpoint of the readiness probe.")
	flag.StringVar(&healthEndpoint, "health-check-endpoint", "healthz", "The endpoint of the health probe.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.Parse()

	cfg := env.GetConfig()
	if len(cfg.Domain) == 0 {
		setupLog.Error(fmt.Errorf("env var DOMAIN should be a non-empty value"), "unable to start manager")
		os.Exit(1)
	}

	// cluster config
	k8sConfig := ctrl.GetConfigOrDie()

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDebugLogs)))
	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		SyncPeriod:             &resyncPeriod,
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
		mgr.GetEventRecorderFor("eventing-controller"),
		cfg,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup the Subscription controller")
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
