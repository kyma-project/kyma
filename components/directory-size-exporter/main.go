package main

import (
	"flag"
	"net/http"

	"github.com/kyma-project/kyma/components/directory-size-exporter/internal/exporter"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var logFormat string
	var logLevel string

	var storagePath string
	var dirsSizeMetricName string
	var port string
	var interval int

	flag.StringVar(&logFormat, "log-format", "text", "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")

	flag.StringVar(&storagePath, "storage-path", "/data/log/flb-storage/", "Path to the data folder we observe")
	flag.StringVar(&dirsSizeMetricName, "metric-name", "telemetry_fsbuffer_usage_bytes", "Buffer size prometheus metric name")
	flag.StringVar(&port, "port", "2021", "Application port name")
	flag.IntVar(&interval, "interval", 30, "Interval with which we reord our metrics")

	exporterLogger, err := logger.New(logger.Format(logFormat), logger.Level(logLevel))
	if err != nil {
		panic(err)
	}

	exp := exporter.NewExporter(storagePath, dirsSizeMetricName)
	exporterLogger.WithContext().Info("Exporter is initialized")

	exp.RecordMetrics(interval)
	exporterLogger.WithContext().Info("Started recording metrics")

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
	exporterLogger.WithContext().Info("Listening on port '" + port + "'")
}
