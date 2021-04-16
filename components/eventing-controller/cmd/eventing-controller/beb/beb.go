package beb

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

// Command implementes the Command interface.
type Command struct {
	scheme          *runtime.Scheme
	cfg             env.Config
	enableDebugLogs bool
	metricsAddr     string
	resyncPeriod    time.Duration
	mgr             manager.Manager
}

// NewCommand creates the command for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommand(enableDebugLogs bool, metricsAddr string, resyncPeriod time.Duration) *Command {
	c := &Command{
		scheme:          runtime.NewScheme(),
		cfg:             env.GetConfig(),
		enableDebugLogs: enableDebugLogs,
		metricsAddr:     metricsAddr,
		resyncPeriod:    resyncPeriod,
	}
	_ = clientgoscheme.AddToScheme(scheme)
	_ = eventingv1alpha1.AddToScheme(scheme)
	_ = apigatewayv1alpha1.AddToScheme(scheme)
	return c
}

// Init implements the Command interface and initializes the BEB command.
func (c *Command) Init() error {
	if len(c.cfg.Domain) == 0 {
		return fmt.Errorf("env var DOMAIN must be a non-empty value")
	}
	k8sConfig := ctrl.GetConfigOrDie()
	ctrl.SetLogger(zap.New(zap.UseDevMode(c.enableDebugLogs)))
	// Create the manager.
	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: c.metricsAddr,
		Port:               9443,
		SyncPeriod:         &c.resyncPeriod,
	})
	if err != nil {
		return err
	}
	c.mgr = mgr
	return nil
}

// Start implements the Command interface and starts the manager.
func (c *Command) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	applicationLister := application.NewLister(ctx, dynamicClient)

	if err := subscription.NewReconciler(
		c.mgr.GetClient(),
		applicationLister,
		c.mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		c.mgr.GetEventRecorderFor("eventing-controller"),
		c.cfg,
	).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to setup the Subscription controller: %v", err)
	}

	if err := c.mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}
	return nil
}
