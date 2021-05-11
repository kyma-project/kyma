package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
)

func main() {
	logger := ctrl.Log.WithName("setup")
	scheme := runtime.NewScheme()
	restCfg := ctrl.GetConfigOrDie()

	// Add schemes.
	if err := beb.AddToScheme(scheme); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}
	if err := nats.AddToScheme(scheme); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Parse flags.
	var enableDebugLogs bool
	var metricsAddr string
	var resyncPeriod time.Duration
	var maxReconnects int
	var reconnectWait time.Duration

	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls (BEB).")
	flag.IntVar(&maxReconnects, "max-reconnects", 10, "Maximum number of reconnect attempts (NATS).")
	flag.DurationVar(&reconnectWait, "reconnect-wait", time.Second, "Wait time between reconnect attempts (NATS).")
	flag.Parse()

	// Init the manager.
	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDebugLogs)))

	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		SyncPeriod:         &resyncPeriod, // CHECK Only used in BEB so far.
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Start the backend manager.
	backendReconciler := &backend.BackendReconciler{
		Client: mgr.GetClient(),
		Cache: mgr.GetCache(),
		Log:    ctrl.Log.WithName("reconciler").WithName("backend"),
	}
	if err := backendReconciler.SetupWithManager(mgr); err != nil {
		logger.Error(err, "unable to start the backend controller")
		os.Exit(1)
	}

	// Instantiate and initialize configured commander.
	var commander commander.Commander

	backend, err := env.Backend()
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	switch backend {
	case env.BACKEND_VALUE_BEB:
		commander = beb.NewCommander(restCfg, enableDebugLogs, metricsAddr, resyncPeriod)
	case env.BACKEND_VALUE_NATS:
		commander = nats.NewCommander(restCfg, enableDebugLogs, metricsAddr, maxReconnects, reconnectWait)
	}

	if err := commander.Init(mgr); err != nil {
		logger.Error(err, "unable to init commander")
		os.Exit(1)
	}

	// Start the commander.
	// TODO Has to be done by backand management controller later.
	logger.Info(fmt.Sprintf("starting %s commander", backend))

	if err := commander.Start(); err != nil {
		logger.Error(err, "unable to start commander")
		os.Exit(1)
	}

	// Start the manager.
	logger.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
