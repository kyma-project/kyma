package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/cmd/eventing-controller/beb"
	"github.com/kyma-project/kyma/components/eventing-controller/cmd/eventing-controller/nats"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
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
	var enableDebugLogs bool
	var metricsAddr string
	var probeAddr string
	var readyEndpoint string
	var healthEndpoint string
	var resyncPeriod time.Duration
	var maxReconnects int
	var reconnectWait time.Duration

	flag.BoolVar(&enableDebugLogs, "enable-debug-logs", false, "Enable debug logs.")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The TCP address that the controller should bind to for serving prometheus metrics.")
	flag.StringVar(&probeAddr, "health-probe-bind-addr", ":8081", "The TCP address that the controller should bind to for serving health probes.")
	flag.StringVar(&readyEndpoint, "ready-check-endpoint", "readyz", "The endpoint of the readiness probe.")
	flag.StringVar(&healthEndpoint, "health-check-endpoint", "healthz", "The endpoint of the health probe.")
	flag.DurationVar(&resyncPeriod, "reconcile-period", time.Minute*10, "Period between triggering of reconciling calls (BEB).")
	flag.IntVar(&maxReconnects, "max-reconnects", 10, "Maximum number of reconnect attempts (NATS).")
	flag.DurationVar(&reconnectWait, "reconnect-wait", time.Second, "Wait time between reconnect attempts (NATS).")
	flag.Parse()

	// Instantiate configured commander.
	var commander Commander

	backend, err := env.Backend()
	if err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	switch backend {
	case env.BACKEND_VALUE_BEB:
		commander = beb.NewCommander(enableDebugLogs, metricsAddr, probeAddr, readyEndpoint, healthEndpoint, resyncPeriod)
	case env.BACKEND_VALUE_NATS:
		commander = nats.NewCommander(enableDebugLogs, metricsAddr, probeAddr, readyEndpoint, healthEndpoint, maxReconnects, reconnectWait)
	}

	// Init and start the commander.
	if err := commander.Init(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("starting %s subscription controller and manager", backend))

	if err := commander.Start(); err != nil {
		logger.Error(err, "unable to start manager")
		os.Exit(1)
	}
}
