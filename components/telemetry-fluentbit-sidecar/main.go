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

var (
	fsBuffeLabelsVector = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "telemetry_fsbuffer_vector",
		Help:      "Disk size for different emitters",
	}, []string{"name"})

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
		panic(err)
	}
}

func recordingIteration(ticker *time.Ticker, quit chan struct{}) {
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
		for _, dir := range directories {
			fsBuffeLabelsVector.WithLabelValues(dir.name).Set(float64(dir.size))
		}
	case <-quit:
		ticker.Stop()
		return
	}
}

func recordMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			recordingIteration(ticker, quit)
		}
	}()
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
	if err != nil {
		return directories, err
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		size, err := dirSize(path + "/" + file.Name())
		if err != nil {
			return directories, err
		}
		directories = append(directories, directory{file.Name(), size})
	}
	return directories, err
}
