package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/cmd/eventing-controller/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/cmd/eventing-controller/nats"
)

const (
	BEBBackend  = "BEB"
	NATSBackend = "NATS"
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

	// Parse flags.
	var backend string
	var enableDebugLogs bool
	var metricsAddr string
	var resyncPeriod time.Duration
	var maxReconnects int
	var reconnectWait time.Duration

	flag.StringVar(&backend, "backend", "nats", "The controller eventing backend NATS or BEB.")
	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls (BEB).")
	flag.IntVar(&maxReconnects, "max-reconnects", 10, "Maximum number of reconnect attempts (NATS).")
	flag.DurationVar(&reconnectWait, "reconnect-wait", time.Second, "Wait time between reconnect attempts (NATS).")
	flag.Parse()

	// Instantiate configured commander.
	var commander Commander

	backend = strings.ToUpper(backend)

	switch backend {
	case BEBBackend:
		commander = beb.NewCommander(enableDebugLogs, metricsAddr, resyncPeriod)
	case NATSBackend:
		commander = nats.NewCommander(enableDebugLogs, metricsAddr, maxReconnects, reconnectWait)
	default:
		logger.Error(fmt.Errorf("specified invalid eventing controller backend: %v", backend), "unable to start manager")
		os.Exit(1)
	}

	// Init and start the commander.
	if err := commander.Init(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("starting %v subscription controller and manager", commander))

	if err := commander.Start(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
