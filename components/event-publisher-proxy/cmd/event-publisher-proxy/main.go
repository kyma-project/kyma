package main

import (
	golog "log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy/beb"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/cmd/event-publisher-proxy/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/latency"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
)

const (
	backendBEB  = "beb"
	backendNATS = "nats"
)

type Config struct {
	// Backend used for Eventing. It could be "nats" or "beb".
	Backend string `envconfig:"BACKEND" required:"true"`

	// AppLogFormat defines the log format.
	AppLogFormat string `envconfig:"APP_LOG_FORMAT" default:"json"`

	// AppLogLevel defines the log level.
	AppLogLevel string `envconfig:"APP_LOG_LEVEL" default:"info"`
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
	opts := options.ParseArgs()

	// parse the config for main:
	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		golog.Fatalf("Failed to read configuration, error: %v", err)
	}

	// init the logger
	logLevel, ok := os.LookupEnv("APP_LOG_LEVEL")
	if !ok {
		golog.Fatal("Missing APP_LOG_LEVEL environment variable")
	}
	logger, err := kymalogger.New(cfg.AppLogFormat, logLevel)
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
	var commander Commander
	switch cfg.Backend {
	case backendBEB:
		commander = beb.NewCommander(opts, metricsCollector, logger)
	case backendNATS:
		commander = nats.NewCommander(opts, metricsCollector, logger)
	default:
		setupLogger.Fatalf("Invalid publisher backend: %v", cfg.Backend)
	}

	// Init the commander.
	if err := commander.Init(); err != nil {
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
	if err := commander.Start(); err != nil {
		setupLogger.Fatalw("Failed to to start publisher", "error", err)
	}

	setupLogger.Info("Shutdown the Event Publisher")
}
