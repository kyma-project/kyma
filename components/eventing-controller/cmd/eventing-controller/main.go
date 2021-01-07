package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
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
	var resyncPeriod time.Duration
	var enableDebugLogs bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.Parse()

	cfg := env.GetConfig()
	if len(cfg.Domain) == 0 {
		setupLog.Error(fmt.Errorf("env var DOMAIN should be a non-empty value"), "unable to start manager")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDebugLogs)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		SyncPeriod:         &resyncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := subscription.NewReconciler(
		mgr.GetClient(),
		mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		mgr.GetEventRecorderFor("eventing-controller"),
		cfg,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup the Subscription controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
