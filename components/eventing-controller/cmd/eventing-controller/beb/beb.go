package beb

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

// Commander implements the Commander interface.
type Commander struct {
	scheme          *runtime.Scheme
	envCfg          env.Config
	restCfg         *rest.Config
	enableDebugLogs bool
	metricsAddr     string
	probeAddr       string
	readyEndpoint   string
	healthEndpoint  string
	resyncPeriod    time.Duration
	mgr             manager.Manager
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(enableDebugLogs bool, metricsAddr, probeAddr, readyEndpoint, healthEndpoint string, resyncPeriod time.Duration) *Commander {
	return &Commander{
		scheme:          runtime.NewScheme(),
		envCfg:          env.GetConfig(),
		enableDebugLogs: enableDebugLogs,
		metricsAddr:     metricsAddr,
		probeAddr:       probeAddr,
		readyEndpoint:   readyEndpoint,
		healthEndpoint:  healthEndpoint,
		resyncPeriod:    resyncPeriod,
	}
}

// Init implements the Commander interface and initializes the BEB command.
func (c *Commander) Init() error {
	// Adding schemas.
	if err := clientgoscheme.AddToScheme(c.scheme); err != nil {
		return err
	}
	if err := eventingv1alpha1.AddToScheme(c.scheme); err != nil {
		return err
	}
	if err := apigatewayv1alpha1.AddToScheme(c.scheme); err != nil {
		return err
	}
	// Check config.
	if len(c.envCfg.Domain) == 0 {
		return fmt.Errorf("env var DOMAIN must be a non-empty value")
	}
	c.restCfg = ctrl.GetConfigOrDie()
	ctrl.SetLogger(zap.New(zap.UseDevMode(c.enableDebugLogs)))
	// Create the manager.
	mgr, err := ctrl.NewManager(c.restCfg, ctrl.Options{
		Scheme:                 c.scheme,
		MetricsBindAddress:     c.metricsAddr,
		HealthProbeBindAddress: c.probeAddr,
		Port:                   9443,
		SyncPeriod:             &c.resyncPeriod,
	})
	if err != nil {
		return err
	}
	c.mgr = mgr
	return nil
}

// Start implements the Commander interface and starts the manager.
func (c *Commander) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dynamicClient := dynamic.NewForConfigOrDie(c.restCfg)
	applicationLister := application.NewLister(ctx, dynamicClient)

	if err := subscription.NewReconciler(
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		c.mgr.GetEventRecorderFor("eventing-controller"), // TODO Harmonization? Add "-beb"?
		c.envCfg,
	).SetupWithManager(c.mgr); err != nil {
		return fmt.Errorf("unable to setup the BEB Subscription Controller: %v", err)
	}

	if err := c.mgr.AddHealthzCheck(c.healthEndpoint, healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %v", err)
	}
	if err := c.mgr.AddReadyzCheck(c.readyEndpoint, healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %v", err)
	}

	if err := c.mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}
	return nil
}
