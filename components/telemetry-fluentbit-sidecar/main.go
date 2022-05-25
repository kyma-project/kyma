package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const logPath = "/data/log/flb-storage/" // pass as flag and check dockerfile entrypoint vs run

func recordMetrics() {
	ticker := time.NewTicker(2 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				size, err := dirSize(logPath)
				if err != nil {
					panic(err)
				}
				opsProcessed.Set(float64(size))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

var (
	opsProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        "fluentbit_fsbuffer_size",
		Help:        "The total disk size of the fluentbit chunk buffer",
		ConstLabels: nil,
	})
)

func main() {
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2021", nil)
	if err != nil {
		return
	}
}

func dirSize(path string) (int64, error) {
	return 60, nil
	//var size int64
	//err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//	if !info.IsDir() {
	//		size += info.Size()
	//	}
	//	return err
	//})
	//return size, err
}
