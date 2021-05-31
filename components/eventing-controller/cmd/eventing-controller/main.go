package main

import (
	"context"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/zapr"

	"github.com/kyma-project/kyma/components/eventing-controller/log"
	"github.com/kyma-project/kyma/components/eventing-controller/options"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/backend"
)

func main() {
	setupLogger := ctrl.Log.WithName("setup")

	opts := new(options.Options)
	if err := opts.Parse(); err != nil {
		setupLogger.Error(err, "failed to parse options")
		os.Exit(1)
	}

	logger, err := log.NewLogger(opts.LogFormat, opts.LogLevel)
	if err != nil {
		setupLogger.Error(err, "failed to initialize logger")
		os.Exit(1)
	}
	defer func() {
		if err := logger.WithContext().Sync(); err != nil {
			setupLogger.Error(err, "failed to flush logger")
		}
	}()

	// set controller core logger
	ctrl.SetLogger(zapr.NewLogger(logger.WithContext().Desugar()))

	// Add schemes.
	scheme := runtime.NewScheme()
	if err := beb.AddToScheme(scheme); err != nil {
		setupLogger.Error(err, "unable to start manager")
		os.Exit(1)
	}
	if err := nats.AddToScheme(scheme); err != nil {
		setupLogger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Init the manager.
	restCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: opts.MetricsAddr,
		Port:               9443,
		SyncPeriod:         &opts.ReconcilePeriod, // CHECK Only used in BEB so far.
	})
	if err != nil {
		setupLogger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Instantiate and initialize all the subscription commanders.
	natsCommander := nats.NewCommander(restCfg, opts.MetricsAddr, opts.MaxReconnects, opts.ReconnectWait)
	if err := natsCommander.Init(mgr); err != nil {
		setupLogger.Error(err, "unable to initialize the NATS commander")
		os.Exit(1)
	}

	bebCommander := beb.NewCommander(restCfg, opts.MetricsAddr, opts.ReconcilePeriod)
	if err := bebCommander.Init(mgr); err != nil {
		setupLogger.Error(err, "unable to initialize the BEB commander")
		os.Exit(1)
	}

	// Start the backend manager.
	ctx := context.Background()
	recorder := mgr.GetEventRecorderFor("backend-controller")
	backendReconciler := backend.NewReconciler(ctx, natsCommander, bebCommander, mgr.GetClient(), mgr.GetCache(), logger, recorder)
	if err := backendReconciler.SetupWithManager(mgr); err != nil {
		setupLogger.Error(err, "unable to start the backend controller")
		os.Exit(1)
	}

	// Start the manager.
	logger.WithContext().With("options", opts).Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLogger.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
