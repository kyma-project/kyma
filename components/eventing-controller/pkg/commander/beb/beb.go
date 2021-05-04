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
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/application"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription"
)

// AddToScheme adds the own schemes to the runtime scheme.
func AddToScheme(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := apigatewayv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

// Commander implements the Commander interface.
type Commander struct {
	envCfg          env.Config
	restCfg         *rest.Config
	enableDebugLogs bool
	metricsAddr     string
	resyncPeriod    time.Duration
	mgr             manager.Manager
}

// NewCommander creates the Commander for BEB and initializes it as far as it
// does not depend on non-common options.
func NewCommander(restCfg *rest.Config, enableDebugLogs bool, metricsAddr string, resyncPeriod time.Duration) *Commander {
	return &Commander{
		envCfg:          env.GetConfig(),
		restCfg:         restCfg,
		enableDebugLogs: enableDebugLogs,
		metricsAddr:     metricsAddr,
		resyncPeriod:    resyncPeriod,
	}
}

// Start implements the Commander interface and starts the manager.
func (c *Commander) Start(mgr manager.Manager) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(c.envCfg.Domain) == 0 {
		return fmt.Errorf("env var DOMAIN must be a non-empty value")
	}
	c.mgr = mgr

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

	if err := c.mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}
	return nil
}

// Stop implements the Commander interface and stops the commander.
func (c *Commander) Stop() error {
	return nil // TODO Stopping and cleanup.
}
