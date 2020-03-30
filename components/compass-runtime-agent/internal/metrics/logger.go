package metrics

import (
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Logger interface {
	Start(quitChannel <-chan struct{}) error
}

type logger struct {
	loggingTimeInterval time.Duration
	resourcesFetcher    ResourcesFetcher
	metricsFetcher      MetricsFetcher
	volumesFetcher      VolumesFetcher
}

func NewMetricsLogger(
	resourcesClientset kubernetes.Interface,
	metricsClientset clientset.Interface,
	loggingTimeInterval time.Duration) Logger {

	return &logger{
		loggingTimeInterval: loggingTimeInterval,
		resourcesFetcher:    newResourcesFetcher(resourcesClientset),
		metricsFetcher:      newMetricsFetcher(metricsClientset),
		volumesFetcher:      newVolumesFetcher(resourcesClientset),
	}
}

func (l *logger) Start(quitChannel <-chan struct{}) error {
	for {
		select {
		case <-time.Tick(l.loggingTimeInterval):
			l.log()
		case <-quitChannel:
			log.Info("Logging stopped.")
			return nil
		}
	}
}

func (l *logger) log() {
	clusterInfo, err := l.fetchClusterInfo()
	if err != nil {
		log.Error(errors.Wrap(err, "failed to fetch cluster info"))
		return
	}

	l.printLogs(clusterInfo)
}

func (l *logger) fetchClusterInfo() (ClusterInfo, error) {
	resources, err := l.resourcesFetcher.FetchNodesResources()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "failed to fetch nodes resources")
	}

	metrics, err := l.metricsFetcher.FetchNodeMetrics()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "failed to fetch nodes metrics")
	}

	volumes, err := l.volumesFetcher.FetchPersistentVolumesCapacity()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "failed to fetch persistent volumes capacity")
	}

	return ClusterInfo{
		Resources: resources,
		Usage:     metrics,
		Volumes:   volumes,
	}, nil
}

func (l *logger) printLogs(clusterInfo ClusterInfo) {
	log.SetFormatter(&log.JSONFormatter{})

	log.WithFields(log.Fields{
		"clusterInfo": clusterInfo,
		"metrics":     true,
	}).Info("Cluster metrics logged successfully.")

	log.SetFormatter(&log.TextFormatter{})
}
