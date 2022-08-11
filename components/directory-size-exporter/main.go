package main

import (
	"directory-size-exporter/exporter"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func readEnvironmentVariable(name string) (string, error) {
	environmentValue := os.Getenv(name)

	if environmentValue == "" {
		return "", fmt.Errorf("You have to define %s environment variable", name)
	}

	return environmentValue, nil
}

var (
	logFormat string
	logLevel  string
)

func main() {
	flag.StringVar(&logFormat, "log-format", "text", "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")

	exporterLogger, err := logger.New(logger.Format(logFormat), logger.Level(logLevel))
	if err != nil {
		panic(err)
	}

	logPath, err := readEnvironmentVariable("STORAGE_PATH")
	if err != nil {
		exporterLogger.WithContext().Error("Error occurred during an attempt to read STORAGE_PATH variable!")
		panic(err)
	}
	exporterLogger.WithContext().Info("Read STORAGE_PATH environment variable")

	dirsSizeMetricName, err := readEnvironmentVariable("DIRECTORIES_SIZE_METRIC")
	if err != nil {
		exporterLogger.WithContext().Error("Error occurred during an attempt to read DIRECTORIES_SIZE_METRIC variable!")
		panic(err)
	}
	exporterLogger.WithContext().Info("Read DIRECTORIES_SIZE_METRIC environment variable")

	port, err := readEnvironmentVariable("METRICS_PORT")
	if err != nil {
		exporterLogger.WithContext().Error("Error occurred during an attempt to read METRICS_PORT variable!")
		panic(err)
	}
	exporterLogger.WithContext().Info("Read METRICS_PORT environment variable")

	exp := exporter.NewExporter(logPath, dirsSizeMetricName)
	exporterLogger.WithContext().Info("Exporter is initialized")

	exp.RecordMetrics()
	exporterLogger.WithContext().Info("Started recording metrics")

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
	exporterLogger.WithContext().Info("Listening port '" + port + "'")
}
