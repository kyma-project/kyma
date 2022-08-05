package main

import (
	"fmt"
	"net/http"
	"os"
	"telemetry-fluentbit-sidecar/exporter"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func readEnvironmentVariable(name string) (string, error) {
	environmentValue := os.Getenv(name)

	if environmentValue == "" {
		return "", fmt.Errorf("You have to define %s environment variable", name)
	}

	return environmentValue, nil
}

func main() {
	logPath, err := readEnvironmentVariable("STORAGE_PATH")
	if err != nil {
		panic(err)
	}

	dirsSizeMetricName, err := readEnvironmentVariable("DIRECTORIES_SIZE_METRIC")
	if err != nil {
		panic(err)
	}

	totalSizeMetricName, err := readEnvironmentVariable("TOTAL_SIZE_METRIC")
	if err != nil {
		panic(err)
	}

	exp := exporter.NewExporter(logPath, dirsSizeMetricName, totalSizeMetricName)

	exp.RecordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":2021", nil)
	if err != nil {
		panic(err)
	}
}
