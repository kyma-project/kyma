package main

import (
	"flag"
	"os"
	"strings"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Commander defines the interface of different implementations
type Commander interface {
	// Init allows main() to pass flag values to the commander instance.
	Init() error

	// Start runs the initialized commander instance.
	Start() error
}

func main() {
	logger := ctrl.Log.WithName("setup")

	var backend string
	var metricsAddr string
	var resyncPeriod time.Duration
	var enableDebugLogs bool

	flag.StringVar(&backend, "backend", "nats", "The controller eventing backend (NATS / BEB).")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.Parse()

	var commander Commander

	switch strings.ToLower(backend) {
	case "nats":
		// TODO Create NATS Commander.
	case "beb":
		commander = beb.NewCommander(enableDebugLogs, metricsAddr, resyncPeriod)
	}

	if err := commander.Init(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	logger.Info("starting Subscription controller and manager")

	if err := cmmander.Start(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
