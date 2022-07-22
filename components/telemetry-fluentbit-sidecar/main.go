package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type directory struct {
	name string
	size int64
}

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
				directories, errDirList := listDirs(logPath)
				if errDirList != nil {
					panic(errDirList)
				}
				for i, dir := range directories {
					fsbufferLabels[i].Set(float64(dir.size))
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

var (
	fsbufferLabels = []prometheus.Gauge{
		promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "telemetry_fsbuffer_size_emitter4",
			Help:        "The emitter.4 size of the fluentbit chunk buffer",
			ConstLabels: nil,
		}),
		promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "telemetry_fsbuffer_size_emitter3",
			Help:        "The emitter.3 size of the fluentbit chunk buffer",
			ConstLabels: nil,
		}),
		promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "telemetry_fsbuffer_size_emitter2",
			Help:        "The emitter.2 size of the fluentbit chunk buffer",
			ConstLabels: nil,
		}),
		promauto.NewGauge(prometheus.GaugeOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "telemetry_fsbuffer_size_tail0",
			Help:        "The tail.0 size of the fluentbit chunk buffer",
			ConstLabels: nil,
		}),
	}

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

func listDirs(path string) ([]directory, error) {
	directories := make([]directory, 0)
	files, err := ioutil.ReadDir(path)
	for _, file := range files {
		if err != nil {
			return directories, err
		}
		if file.IsDir() {
			size, innerErr := dirSize(path + "/" + file.Name())
			if innerErr != nil {
				return directories, innerErr
			}
			directories = append(directories, directory{file.Name(), size})
		}
	}
	return directories, err
} // iterate through dir, for every subfolder calls dirSize
