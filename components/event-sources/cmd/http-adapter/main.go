package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"
	"github.com/kelseyhightower/envconfig"
	"go.opencensus.io/stats/view"
	"go.uber.org/zap"
	"knative.dev/eventing/pkg/adapter"
	"knative.dev/eventing/pkg/kncloudevents"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/profiling"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/source"
	pkgtracing "knative.dev/pkg/tracing"

	eshttp "github.com/kyma-project/kyma/components/event-sources/adapter/http"
)

const (
	defaultMaxIdleConnections        = 1000
	defaultMaxIdleConnectionsPerHost = 1000
)

func main() {
	setupAdapter("http-source", eshttp.NewEnvConfig, eshttp.NewAdapter)
}

func setupAdapter(component string, ector adapter.EnvConfigConstructor, ctor adapter.AdapterConstructor) {
	flag.Parse()

	ctx := signals.NewContext()

	env := ector()
	if err := envconfig.Process("", env); err != nil {
		log.Fatalf("Error processing env var: %s", err)
	}

	// Convert json logging.Config to logging.Config.
	loggingConfig, err := logging.JsonToLoggingConfig(env.GetLoggingConfigJson())
	if err != nil {
		fmt.Printf("[ERROR] failed to process logging config: %s", err.Error())
		// Use default logging config.
		if loggingConfig, err = logging.NewConfigFromMap(map[string]string{}); err != nil {
			// If this fails, there is no recovering.
			panic(err)
		}
	}

	logger, _ := logging.NewLoggerFromConfig(loggingConfig, component)
	defer flush(logger)
	ctx = logging.WithLogger(ctx, logger)

	// Report stats on Go memory usage every 30 seconds.
	msp := metrics.NewMemStatsAll()
	msp.Start(ctx, 30*time.Second)
	if err := view.Register(msp.DefaultViews()...); err != nil {
		logger.Fatal("Error exporting go memstats view: %v", zap.Error(err))
	}

	// Convert json metrics.ExporterOptions to metrics.ExporterOptions.
	metricsConfig, err := metrics.JsonToMetricsOptions(env.GetMetricsConfigJson())
	if err != nil {
		logger.Error("failed to process metrics options", zap.Error(err))
	} else {
		if err := metrics.UpdateExporter(*metricsConfig, logger); err != nil {
			logger.Error("failed to create the metrics exporter", zap.Error(err))
		}
	}

	// Check if metrics config contains profiling flag
	if metricsConfig != nil && metricsConfig.ConfigMap != nil {
		if enabled, err := profiling.ReadProfilingFlag(metricsConfig.ConfigMap); err == nil {
			if enabled {
				// Start a goroutine to server profiling metrics
				logger.Info("Profiling enabled")
				go func() {
					server := profiling.NewServer(profiling.NewHandler(logger, true))
					// Don't forward ErrServerClosed as that indicates we're already shutting down.
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logger.Error("profiling server failed", zap.Error(err))
					}
				}()
			}
		} else {
			logger.Error("error while reading profiling flag", zap.Error(err))
		}
	}

	reporter, err := source.NewStatsReporter()
	if err != nil {
		logger.Error("error building statsreporter", zap.Error(err))
	}

	options := []cloudeventshttp.Option{
		cloudevents.WithBinaryEncoding(),
		cloudevents.WithMiddleware(pkgtracing.HTTPSpanMiddleware),
	}

	switch v := env.(type) {
	case eshttp.AdapterEnvConfigAccessor:
		options = append(options, cloudevents.WithPort(v.GetPort()))
		options = append(options, cloudevents.WithPath(eshttp.EndpointCE))
		options = append(options, cloudevents.WithMiddleware(eshttp.WithReadinessMiddleware))
	default:
		logger.Infof("Wrong ector type %v", v)
		logger.Info("Tracing disabled")
	}

	httpTransport, err := cloudevents.NewHTTPTransport(
		options...,
	)
	if err != nil {
		logger.Fatal("Unable to create CE transport", zap.Error(err))
	}

	connectionArgs := kncloudevents.ConnectionArgs{
		MaxIdleConns:        defaultMaxIdleConnections,
		MaxIdleConnsPerHost: defaultMaxIdleConnectionsPerHost,
	}

	ceClient, err := kncloudevents.NewDefaultClientGivenHttpTransport(
		httpTransport,
		&connectionArgs)

	if err != nil {
		logger.Fatal("error building cloud event client", zap.Error(err))
	}

	// Configuring the adapter
	adptr := ctor(ctx, env, ceClient, reporter)

	logger.Info("Starting Receive Adapter", zap.Any("adapter", adptr))

	if err := adptr.Start(ctx.Done()); err != nil {
		logger.Warn("start returned an error", zap.Error(err))
	}
}

func flush(logger *zap.SugaredLogger) {
	_ = logger.Sync()
	metrics.FlushExporter()
}
