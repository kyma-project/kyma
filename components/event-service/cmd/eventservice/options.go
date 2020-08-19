package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	requestTimeout  int
	sourceID        string
	requestLogging  bool
	maxRequestSize  int64
	eventMeshURL    string
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	sourceID := flag.String("sourceId", "stage.local.kyma.commerce", "The source id of the events")
	maxRequestSize := flag.Int64("maxRequestSize", 65536, "The maximum request size in bytes")
	eventMeshURL := flag.String("eventMeshURL", "http://localhost:9001/events", "Target URL for the HTTP Source Adapter, Entrypoint to event mesh")

	flag.Parse()

	return &options{
		externalAPIPort: *externalAPIPort,
		requestTimeout:  *requestTimeout,
		requestLogging:  *requestLogging,
		sourceID:        *sourceID,
		maxRequestSize:  *maxRequestSize,
		eventMeshURL:    *eventMeshURL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --eventMeshURL=%s --requestTimeout=%d --sourceId=%s --requestLogging=%t --maxRequestSize=%d",
		o.externalAPIPort, o.eventMeshURL, o.requestTimeout, o.sourceID, o.requestLogging, o.maxRequestSize)
}
