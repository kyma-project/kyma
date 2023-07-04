package main

import (
	golog "log"

	"github.com/kelseyhightower/envconfig"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander/eventmesh"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/latency"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
)

const (
	backendEventMesh = "beb"
	backendNATS      = "nats"
)

type Config struct {
	// Backend used for Eventing. It could be "nats" or "beb".
	Backend string `envconfig:"BACKEND" required:"true"`

	// AppLogFormat defines the log format.
	AppLogFormat string `envconfig:"APP_LOG_FORMAT" default:"json"`

	// AppLogLevel defines the log level.
	AppLogLevel string `envconfig:"APP_LOG_LEVEL" default:"info"`
}

func main() {
	opts := options.New()
	if err := opts.Parse(); err != nil {
		golog.Fatalf("Failed to parse options, error: %v", err)
	}

	// parse the config for main:
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		golog.Fatalf("Failed to read configuration, error: %v", err)
	}

	// init the logger
	logger, err := kymalogger.New(cfg.AppLogFormat, cfg.AppLogLevel)
	if err != nil {
		golog.Fatalf("Failed to initialize logger, error: %v", err)
	}
	defer func() {
		if err := logger.WithContext().Sync(); err != nil {
			golog.Printf("Failed to flush logger, error: %v", err)
		}
	}()
	setupLogger := logger.WithContext().With("backend", cfg.Backend)

	// metrics collector
	metricsCollector := metrics.NewCollector(latency.NewBucketsProvider())
	prometheus.MustRegister(metricsCollector)

	// Instantiate configured commander.
	var c commander.Commander
	switch cfg.Backend {
	case backendEventMesh:
		c = eventmesh.NewCommander(opts, metricsCollector, logger)
	case backendNATS:
		c = nats.NewCommander(opts, metricsCollector, logger)
	default:
		setupLogger.Fatalf("Invalid publisher backend: %v", cfg.Backend)
	}

	// Init the commander.
	if err := c.Init(); err != nil {
		setupLogger.Fatalw("Commander initialization failed", "error", err)
	}

	// Start the metrics server.
	metricsServer := metrics.NewServer(logger)
	defer metricsServer.Stop()
	if err := metricsServer.Start(opts.MetricsAddress); err != nil {
		setupLogger.Infow("Failed to start metrics server", "error", err)
	}

	setupLogger.Infof("Starting publisher to: %v", cfg.Backend)

	// Start the commander.
	if err := c.Start(); err != nil {
		setupLogger.Fatalw("Failed to to start publisher", "error", err)
	}

	setupLogger.Info("Shutdown the Event Publisher")
}
