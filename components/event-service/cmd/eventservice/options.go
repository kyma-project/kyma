package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	eventsTargetURL string
	requestTimeout  int
	sourceID        string
	requestLogging  bool
	maxRequestSize  int64
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	eventsTargetURL := flag.String("eventsTargetURL", "http://localhost:9001/events", "Target URL for events to be sent.")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	sourceID := flag.String("sourceId", "stage.local.kyma.commerce", "The source id of the events")
	maxRequestSize := flag.Int64("maxRequestSize", 65536, "The maximum request size in bytes")

	flag.Parse()

	return &options{
		externalAPIPort: *externalAPIPort,
		eventsTargetURL: *eventsTargetURL,
		requestTimeout:  *requestTimeout,
		requestLogging:  *requestLogging,
		sourceID:        *sourceID,
		maxRequestSize:  *maxRequestSize,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --eventsTargetURL=%s --requestTimeout=%d --sourceId=%s --requestLogging=%t --maxRequestSize=%d",
		o.externalAPIPort, o.eventsTargetURL, o.requestTimeout, o.sourceID, o.requestLogging, o.maxRequestSize)
}
