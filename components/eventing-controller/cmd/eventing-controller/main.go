package main

import (
	"context"
	"os"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"github.com/go-logr/zapr"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/jetstream"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/kyma-project/kyma/components/eventing-controller/controllers/backend"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/options"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/nats"
)

func main() {
	setupLogger := ctrl.Log.WithName("setup")

	opts := options.New()
	if err := opts.Parse(); err != nil {
		setupLogger.Error(err, "parse options failed")
		os.Exit(1)
	}

	ctrLogger, err := logger.New(opts.LogFormat, opts.LogLevel)
	if err != nil {
		setupLogger.Error(err, "initialize logger failed")
		os.Exit(1)
	}
	defer func() {
		if err := ctrLogger.WithContext().Sync(); err != nil {
			setupLogger.Error(err, "flush logger failed")
		}
	}()

	// Set controller core logger.
	ctrl.SetLogger(zapr.NewLogger(ctrLogger.WithContext().Desugar()))

	// Instantiate and initialize all the subscription managers.
	restCfg := ctrl.GetConfigOrDie()
	scheme := runtime.NewScheme()

	var natsSubMgr subscriptionmanager.Manager
	natsConfig := env.GetNatsConfig(opts.MaxReconnects, opts.ReconnectWait)
	if opts.EnableJetStreamBackend {
		natsSubMgr = jetstream.NewSubscriptionManager(restCfg, natsConfig, opts.MetricsAddr, ctrLogger)
		if err := jetstream.AddToScheme(scheme); err != nil {
			setupLogger.Error(err, "start manager failed")
			os.Exit(1)
		}
	} else {
		natsSubMgr = nats.NewSubscriptionManager(restCfg, natsConfig, opts.MetricsAddr, ctrLogger)
		if err := nats.AddToScheme(scheme); err != nil {
			setupLogger.Error(err, "start manager failed")
			os.Exit(1)
		}
	}

	bebSubMgr := beb.NewSubscriptionManager(restCfg, opts.MetricsAddr, opts.ReconcilePeriod, ctrLogger)
	if err := beb.AddToScheme(scheme); err != nil {
		setupLogger.Error(err, "start manager failed")
		os.Exit(1)
	}

	// Init the manager.
	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     opts.MetricsAddr,
		HealthProbeBindAddress: opts.ProbeAddr,
		Port:                   9443,
		SyncPeriod:             &opts.ReconcilePeriod, // CHECK Only used in BEB so far.
	})
	if err != nil {
		setupLogger.Error(err, "start manager failed")
		os.Exit(1)
	}

	if err := natsSubMgr.Init(mgr); err != nil {
		setupLogger.Error(err, "initialize NATS subscription manager failed")
		os.Exit(1)
	}

	if err := bebSubMgr.Init(mgr); err != nil {
		setupLogger.Error(err, "initialize BEB subscription manager failed")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck(opts.HealthEndpoint, healthz.Ping); err != nil {
		setupLogger.Error(err, "setup health check failed")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck(opts.ReadyEndpoint, healthz.Ping); err != nil {
		setupLogger.Error(err, "setup ready check failed")
		os.Exit(1)
	}

	// Start the backend manager.
	ctx := context.Background()
	recorder := mgr.GetEventRecorderFor("backend-controller")
	backendReconciler := backend.NewReconciler(ctx, natsSubMgr, natsConfig, bebSubMgr, mgr.GetClient(), ctrLogger, recorder)
	if err := backendReconciler.SetupWithManager(mgr); err != nil {
		setupLogger.Error(err, "start backend controller failed")
		os.Exit(1)
	}

	// Start the controller manager.
	ctrLogger.WithContext().With("options", opts).Info("start controller manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLogger.Error(err, "start controller manager failed")
		os.Exit(1)
	}
}
