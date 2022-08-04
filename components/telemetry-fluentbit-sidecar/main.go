package main

import (
	"fmt"
	"net/http"
	"os"
	"telemetry-fluentbit-sidecar/exporter"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// func dataPathVariable() error {
// 	logPath = os.Getenv("STORAGE_PATH")
// 	if logPath == "" {
// 		return fmt.Errorf("You have to define data storage path by setting STORAGE_PATH environment variable")
// 	}
// 	return nil
// }

// func readEnvironmentVariable() {
// 	dirsSizeMetricName := os.Getenv("DIRECTORIES_SIZE_METRIC")
// 	totalSizeMetricName := os.Getenv("TOTAL_SIZE_METRIC")

// 	if dirsSizeMetricName == "" {
// 		return fmt.Errorf("You have to define metric name by setting DIRECTORIES_SIZE_METRIC environment variable")
// 	}

// 	if totalSizeMetricName == "" {
// 		return fmt.Errorf("You have to define metric name by setting TOTAL_SIZE_METRIC environment variable")
// 	}
// }

func readEnvironmentVariable(name string) (string, error) {
	environmentValue := os.Getenv(name)

	if environmentValue == "" {
		return "", fmt.Errorf("You have to define ", name, " environment variable")
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
