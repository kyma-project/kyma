package metrics

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"time"
)

type Logger interface {
	Log(quitChannel <-chan bool)
}

type logger struct {
	loggingTimeInterval time.Duration
	resourcesFetcher    ResourcesFetcher
	metricsFetcher      MetricsFetcher
}

func NewMetricsLogger(
	resourcesClientset kubernetes.Interface,
	metricsClientset clientset.Interface,
	loggingTimeInterval time.Duration) Logger {

	return &logger{
		loggingTimeInterval: loggingTimeInterval,
		resourcesFetcher:    newResourcesFetcher(resourcesClientset),
		metricsFetcher:      newMetricsFetcher(metricsClientset),
	}
}

func (l *logger) Log(quitChannel <-chan bool) {
	for {
		select {
		case <-time.Tick(l.loggingTimeInterval):
			l.log()
		case <-quitChannel:
			log.Info("Logging stopped.")
			return
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

	return ClusterInfo{
		ShouldBeFetched: true,
		Resources:       resources,
		Usage:           metrics,
		Time:            time.Now(),
	}, nil
}

func (l *logger) printLogs(clusterInfo ClusterInfo) {
	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: true,
	})

	log.WithFields(log.Fields{
		"metrics": clusterInfo,
	}).Info("Cluster metrics logged successfully.")

	log.SetFormatter(&log.TextFormatter{})
}
