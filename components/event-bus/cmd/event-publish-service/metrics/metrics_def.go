package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	// Namespace metric label
	Namespace = "namespace"
	// Status metric label
	Status = "status"
	// SourceID metric label
	SourceID = "source_id"
	// EventType metric label
	EventType = "event_type"
	// EventTypeVersion metric label
	EventTypeVersion = "event_type_version"
)

var (
	// TotalPublishedMessages metric
	TotalPublishedMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "total_published_messages",
		Help: "The total number of published messages",
	}, []string{Namespace, Status, SourceID, EventType, EventTypeVersion})
)
