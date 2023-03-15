package main

import (
	golog "log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander/beb"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/commander/nats"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler/health"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/k8s"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/logger"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/metrics/latency"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/options"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/watcher"
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

	appConfig := watcher.New()
	k8sConfig := k8s.ConfigOrDie()
	k8sClient := k8s.ClientOrDie(k8sConfig)
	loggerInstance := logger.New(appConfig, k8sConfig)

	watcher.NewWatcher(k8sClient, logger.Namespace, logger.ConfigMapName).
		OnUpdateNotify(loggerInstance).
		Watch()

	// init the logger
	logger, err := kymalogger.NewWithAtomicLevel(cfg.AppLogFormat, cfg.AppLogLevel)
	if err != nil {
		golog.Fatalf("Failed to initialize logger, error: %v", err)
	}
	defer func() {
		if err := logger.WithContext().Sync(); err != nil {
			golog.Printf("Failed to flush logger, error: %v", err)
		}
	}()
	setupLogger := logger.WithContext().With("backend", cfg.Backend)

	http.HandleFunc(health.ReadinessURI, health.DefaultCheck)
	http.HandleFunc(health.LivenessURI, health.DefaultCheck)
	setupLogger.Infof("Somekind of problem", http.ListenAndServe(appConfig.ServerAddress, nil))

	// metrics collector
	metricsCollector := metrics.NewCollector(latency.NewBucketsProvider())
	prometheus.MustRegister(metricsCollector)

	// Instantiate configured commander.
	var c commander.Commander
	switch cfg.Backend {
	case backendBEB:
		c = beb.NewCommander(opts, metricsCollector, logger)
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
