package main

import (
	"context"
	"log"

	"github.com/go-logr/zapr"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha2"
	"github.com/kyma-project/kyma/components/eventing-controller/controllers/backend"
	"github.com/kyma-project/kyma/components/eventing-controller/internal/featureflags"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/options"
	backendmetrics "github.com/kyma-project/kyma/components/eventing-controller/pkg/backend/metrics"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/eventmesh"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager/jetstream"
)

func main() {
	opts := options.New()
	if err := opts.Parse(); err != nil {
		log.Fatalf("Failed to parse options, error: %v", err)
	}

	ctrLogger, err := logger.New(opts.LogFormat, opts.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger, error: %v", err)
	}
	defer func() {
		if err = ctrLogger.WithContext().Sync(); err != nil {
			log.Printf("Failed to flush logger, error: %v", err)
		}
	}()

	// Set controller core logger.
	ctrl.SetLogger(zapr.NewLogger(ctrLogger.WithContext().Desugar()))

	// prepare the setup logger
	setupLogger := ctrLogger.WithContext().Named("setup")

	// Instantiate and initialize all the subscription managers.
	restCfg := ctrl.GetConfigOrDie()
	scheme := runtime.NewScheme()

	metricsCollector := backendmetrics.NewCollector()
	metricsCollector.RegisterMetrics()

	var natsSubMgr subscriptionmanager.Manager

	natsConfig, err := env.GetNATSConfig(opts.MaxReconnects, opts.ReconnectWait)
	if err != nil {
		setupLogger.Fatalw("Failed to load configuration", "error", err)
	}
	natsSubMgr = jetstream.NewSubscriptionManager(restCfg, natsConfig, opts.MetricsAddr, metricsCollector, ctrLogger)
	if err = jetstream.AddToScheme(scheme); err != nil {
		setupLogger.Fatalw("Failed to start manager", "backend", v1alpha1.NatsBackendType, "error", err)
	}
	if err = jetstream.AddV1Alpha2ToScheme(scheme); err != nil {
		setupLogger.Fatalw("Failed to start manager", "backend", v1alpha1.NatsBackendType, "error", err)
	}

	// Get env config and set feature flags
	envConfig := env.GetConfig()
	featureflags.SetEventingWebhookAuthEnabled(envConfig.EventingWebhookAuthEnabled)
	featureflags.SetNATSProvisioningEnabled(envConfig.NATSProvisioningEnabled)

	bebSubMgr := eventmesh.NewSubscriptionManager(restCfg,
		opts.MetricsAddr,
		opts.ReconcilePeriod,
		ctrLogger,
		metricsCollector)
	if err = eventmesh.AddToScheme(scheme); err != nil {
		setupLogger.Fatalw("Failed to start subscription manager", "backend", v1alpha1.BEBBackendType, "error", err)
	}
	if err = eventmesh.AddV1Alpha2ToScheme(scheme); err != nil {
		setupLogger.Fatalw("Failed to start subscription manager", "backend", v1alpha1.BEBBackendType, "error", err)
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
		setupLogger.Fatalw("Failed to start manager", "error", err)
	}

	if err = natsSubMgr.Init(mgr); err != nil {
		setupLogger.Fatalw("Failed to initialize subscription manager", "backend", v1alpha1.NatsBackendType, "error", err)
	}

	if err = bebSubMgr.Init(mgr); err != nil {
		setupLogger.Fatalw("Failed to initialize subscription manager", "backend", v1alpha1.BEBBackendType, "error", err)
	}

	setupLogger.Infow("Starting the webhook server")

	if err = (&v1alpha1.Subscription{}).SetupWebhookWithManager(mgr); err != nil {
		setupLogger.Fatalw("Failed to create webhook", "error", err)
	}

	if err = (&v1alpha2.Subscription{}).SetupWebhookWithManager(mgr); err != nil {
		setupLogger.Fatalw("Failed to create webhook", "error", err)
	}

	if err = mgr.AddHealthzCheck(opts.HealthEndpoint, healthz.Ping); err != nil {
		setupLogger.Fatalw("Failed to setup health check", "error", err)
	}
	if err = mgr.AddReadyzCheck(opts.ReadyEndpoint, healthz.Ping); err != nil {
		setupLogger.Fatalw("Failed to setup ready check", "error", err)
	}

	// Start the backend manager.
	ctx := context.Background()
	recorder := mgr.GetEventRecorderFor("backend-controller")
	backendConfig := env.GetBackendConfig()
	backendReconciler := backend.NewReconciler(ctx, natsSubMgr, natsConfig, envConfig, backendConfig, bebSubMgr,
		mgr.GetClient(), ctrLogger, recorder)
	if err = backendReconciler.SetupWithManager(mgr); err != nil {
		setupLogger.Fatalw("Failed to start backend controller", "error", err)
	}

	// Start the controller manager.
	ctrLogger.WithContext().With("options", opts).Info("start controller manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLogger.Fatalw("Failed to start controller manager", "error", err)
	}
}
