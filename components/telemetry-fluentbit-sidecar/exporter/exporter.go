package exporter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Exporter interface {
	RecordMetrics()
}

type exporter struct{}

type directory struct {
	name string
	size int64
}

var (
	fsBuffeLabelsVector *prometheus.GaugeVec
	fsbufferSize        prometheus.Gauge
	logPath             string
)

func NewExporter(dataPath string, dirsSizeMetricName string, totalSizeMetricName string) Exporter {
	logPath = dataPath

	inititalizePrometheusMetrics(dirsSizeMetricName, totalSizeMetricName)
	return &exporter{}
}

func (v *exporter) RecordMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				recordingIteration(logPath, ticker)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func inititalizePrometheusMetrics(dirsSizeMetricName string, totalSizeMetricName string) {
	fsBuffeLabelsVector = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "",
		Subsystem: "",
		Name:      dirsSizeMetricName,
		Help:      "Disk size for different emitters",
	}, []string{"name"})

	fsbufferSize = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        totalSizeMetricName,
		Help:        "The total disk size of the fluentbit chunk buffer",
		ConstLabels: nil,
	})
}

func recordingIteration(logPath string, ticker *time.Ticker) {
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
