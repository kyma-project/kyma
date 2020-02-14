package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
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

func NewMetricsLogger(config *rest.Config, loggingTimeInterval time.Duration) (Logger, error) {
	resourcesFetcher, err := newResourcesFetcher(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new resources fetcher")
	}

	metricsFetcher, err := newMetricsFetcher(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new metrics fetcher")
	}

	return &logger{
		loggingTimeInterval: loggingTimeInterval,
		resourcesFetcher:    resourcesFetcher,
		metricsFetcher:      metricsFetcher,
	}, nil
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
	}

	bytes, err := json.Marshal(clusterInfo)
	if err != nil {
		log.Error(errors.Wrap(err, "failed to marshall json"))
	}

	fmt.Println(string(bytes))
	log.Info("Cluster metrics logged successfully.")
}

func (l *logger) fetchClusterInfo() (ClusterInfo, error) {
	resources, err := l.resourcesFetcher.FetchNodesResources()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "failed to fetch nodes resources")
	}

	metrics, err := l.metricsFetcher.FetchNodeMetrics()
	if err != nil {
		return ClusterInfo{}, errors.Wrap(err, "failed to fetch node metrics")
	}

	return ClusterInfo{
		Metrics:   true,
		Resources: resources,
		Usage:     metrics,
		Time:      time.Now(),
	}, nil
}
