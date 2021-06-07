package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy/beb"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	backendBEB  = "beb"
	backendNATS = "nats"
)

type Config struct {
	// Backend used for Eventing. It could be "nats" or "beb"
	Backend string `envconfig:"BACKEND" required:"true"`
}

// Commander defines the interface of different implementations
type Commander interface {
	// Init allows main() to pass flag values to the commander instance.
	Init() error

	// Start runs the initialized commander instance.
	Start() error

	// Stop stops the commander instance.
	Stop() error
}

func main() {
	logger := logrus.New()
	opts := options.ParseArgs()

	// parse the config for main:
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		logger.Fatalf("Read configuration failed with error: %s", err)
	}

	// metrics collector
	metricsCollector := metrics.NewCollector()
	prometheus.MustRegister(metricsCollector)

	// Instantiate configured commander.
	var commander Commander
	switch cfg.Backend {
	case backendBEB:
		commander = beb.NewCommander(opts, metricsCollector, logger)
	case backendNATS:
		commander = nats.NewCommander(opts, metricsCollector, logger)
	default:
		logger.Fatalf("Invalid publisher backend: %v", cfg.Backend)
	}

	// Init the commander.
	if err := commander.Init(); err != nil {
		logger.Fatalf("Initialization failed: %s", err)
	}

	// Start the metrics server.
	metricsServer := metrics.NewServer(logger)
	defer metricsServer.Stop()
	if err := metricsServer.Start(opts.MetricsAddress); err != nil {
		logger.Infof("Metrics server failed to start with error: %v", err)
	}

	logger.Infof("Starting publisher to: %v", cfg.Backend)

	// Start the commander.
	if err := commander.Start(); err != nil {
		logger.Fatalf("Unable to start publisher: %s", err)
	}

	logger.Info("Shutdown the Event Publisher Proxy")
}
