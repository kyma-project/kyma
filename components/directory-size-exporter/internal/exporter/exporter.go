package exporter

import (
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Exporter interface {
	RecordMetrics(interval int)
}

type exporter struct {
	FsBuffeLabelsVector *prometheus.GaugeVec
	LogPath             string
}

type directory struct {
	name string
	size int64
}

func NewExporter(dataPath string, dirsSizeMetricName string) Exporter {
	metricsGague := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "",
		Subsystem: "",
		Name:      dirsSizeMetricName,
		Help:      "Disk size for different emitters",
	}, []string{"directory"})

	return &exporter{
		FsBuffeLabelsVector: metricsGague,
		LogPath:             dataPath,
	}
}

func (v *exporter) RecordMetrics(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				v.recordingIteration(v.LogPath)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (v *exporter) recordingIteration(logPath string) {
	directories, errDirList := listDirs(logPath)
	if errDirList != nil {
		panic(errDirList)
	}
	for _, dir := range directories {
		v.FsBuffeLabelsVector.WithLabelValues(dir.name).Set(float64(dir.size))
	}
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func listDirs(path string) ([]directory, error) {
	directories := make([]directory, 0)
	files, err := os.ReadDir(path)
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
