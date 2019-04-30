package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	Namespace        = "namespace"
	Status           = "status"
	SourceID         = "source_id"
	EventType        = "event_type"
	EventTypeVersion = "event_type_version"
)

var (
	TotalPublishedMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "total_published_messages",
		Help: "The total number of published messages",
	}, []string{Namespace, Status, SourceID, EventType, EventTypeVersion})
)
