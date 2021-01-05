package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	log "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	subscription "github.com/kyma-project/kyma/components/eventing-controller/reconciler/subscription-nats"
)

func main() {
	setupLog := ctrl.Log.WithName("setup")

	var metricsAddr string
	var reSyncPeriod time.Duration
	var enableDebugLogs bool
	var maxReconnects int
	var reconnectWait time.Duration
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&reSyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.IntVar(&maxReconnects, "max-reconnects", 10, "Maximum number of reconnect attempts.")
	flag.DurationVar(&reconnectWait, "reconnect-wait", time.Second, "Wait time between reconnect attempts.")
	flag.Parse()

	cfg := env.GetNatsConfig(maxReconnects, reconnectWait)
	log.Info("Nats config URL: ", cfg.Url)
	if len(cfg.Url) == 0 {
		setupLog.Error(fmt.Errorf("env var URL should be a non-empty value"), "unable to start manager")
		os.Exit(1)
	}

	scheme, err := setupScheme()
	if err != nil {
		setupLog.Error(err, "failed to setup scheme")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(enableDebugLogs)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		SyncPeriod:         &reSyncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := subscription.NewReconciler(
		mgr.GetClient(),
		mgr.GetCache(),
		ctrl.Log.WithName("reconciler").WithName("Subscription"),
		mgr.GetEventRecorderFor("eventing-controller-nats"),
		cfg,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to setup the NATS Subscription controller")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to start manager")
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
