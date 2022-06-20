package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TODO pass as arg or get from env
const logPath = "../data/log/flb-storage/"

func recordMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				size, err := dirSize(logPath)
				if err != nil {
					panic(err)
				}
				fsbufferSize.Set(float64(size))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

var (
	fsbufferSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        "telemetry_fsbuffer_size",
		Help:        "The total disk size of the fluentbit chunk buffer",
		ConstLabels: nil,
	})

	fsbufferMax = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        "telemetry_fsbuffer_limit",
		Help:        "The maximum defined size of the fluentbit chunk buffer",
		ConstLabels: nil,
	})
)

func main() {
	max, err := strconv.ParseFloat(os.Getenv("MAX_FSBUFFER_SIZE"), 64)
	if err != nil {
		panic(err)
	}

	fsbufferMax.Set(max)
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":2021", nil)
	if err != nil {
		return
	}
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
