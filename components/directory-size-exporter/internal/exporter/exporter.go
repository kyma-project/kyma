package exporter

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Exporter interface {
	RecordMetrics(interval int)
}

type exporter struct {
	DirectoriesLabelsVector *prometheus.GaugeVec
	LogPath                 string
	Logger                  *logger.Logger
}

type directory struct {
	name string
	size int64
}

func NewExporter(dataPath string, metricName string, logger *logger.Logger) Exporter {
	metricsGague := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "",
		Subsystem: "",
		Name:      metricName,
		Help:      "Folder sizes of sub-directories",
	}, []string{"directory"})

	return &exporter{
		DirectoriesLabelsVector: metricsGague,
		LogPath:                 dataPath,
		Logger:                  logger,
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
	if errDirList != nil && v.Logger != nil {
		v.Logger.WithContext().Error(errDirList)
	}
	for _, dir := range directories {
		v.DirectoriesLabelsVector.WithLabelValues(dir.name).Set(float64(dir.size))
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
