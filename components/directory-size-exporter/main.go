package main

import (
	"errors"
	"flag"
	"net/http"

	"github.com/kyma-project/kyma/components/directory-size-exporter/internal/exporter"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	storagePath string
	metricName  string
)

func main() {
	var logFormat string
	var logLevel string
	var port string
	var interval int

	flag.StringVar(&logFormat, "log-format", "text", "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")

	flag.StringVar(&storagePath, "storage-path", "", "Path to the observed data folder")
	flag.StringVar(&metricName, "metric-name", "", "Metric name used for exporting the folder size")
	flag.StringVar(&port, "port", "2021", "Port for exposing the metrics")
	flag.IntVar(&interval, "interval", 30, "Interval to calculate the metric ")

	

	flag.Parse()
	if err := validateFlags(); err != nil {
		panic(err)
	}

	exporterLogger, err := logger.New(logger.Format(logFormat), logger.Level(logLevel))
	if err != nil {
		panic(err)
	}

	exp := exporter.NewExporter(storagePath, metricName, exporterLogger)
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

func validateFlags() error {
	if storagePath == "" {
		return errors.New("--storage-path flag is required")
	}
	if metricName == "" {
		return errors.New("--metric-name flag is required")
	}
	return nil
}
