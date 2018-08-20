package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort   int
	eventsTargetURL   string
	requestTimeout    int
	sourceNamespace   string
	sourceType        string
	sourceEnvironment string
	requestLogging    bool
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	eventsTargetURL := flag.String("eventsTargetURL", "http://localhost:9001/events", "Target URL for events to be sent.")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")

	sourceNamespace := flag.String("sourceNamespace", "local.kyma.commerce", "The organization publishing the event.")
	sourceType := flag.String("sourceType", "commerce", "The type of the event source.")
	sourceEnvironment := flag.String("sourceEnvironment", "stage", "The name of the event source environment.")

	flag.Parse()

	return &options{
		externalAPIPort:   *externalAPIPort,
		eventsTargetURL:   *eventsTargetURL,
		requestTimeout:    *requestTimeout,
		requestLogging:    *requestLogging,
		sourceNamespace:   *sourceNamespace,
		sourceType:        *sourceType,
		sourceEnvironment: *sourceEnvironment,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --eventsTargetURL=%s --requestTimeout=%d --sourceNamespace=%s --sourceType=%s --sourceEnvironment=%s --requestLogging=%t",
		o.externalAPIPort, o.eventsTargetURL, o.requestTimeout, o.sourceNamespace, o.sourceType, o.sourceEnvironment, o.requestLogging)
}
