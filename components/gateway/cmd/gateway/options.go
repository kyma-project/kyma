package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort   int
	proxyPort         int
	eventsTargetURL   string
	remoteEnvironment string
	namespace         string
	requestTimeout    int
	skipVerify        bool
	proxyTimeout      int
	sourceNamespace   string
	sourceType        string
	sourceEnvironment string
	requestLogging    bool
	proxyCacheTTL     int
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	eventsTargetURL := flag.String("eventsTargetURL", "http://localhost:9001/events", "Target URL for events to be sent.")
	remoteEnvironment := flag.String("remoteEnvironment", "default-ec", "Remote environment name for reading and updating services.")
	namespace := flag.String("namespace", "kyma-system", "Namespace used by Gateway")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")

	sourceNamespace := flag.String("sourceNamespace", "local.kyma.commerce", "The organization publishing the event.")
	sourceType := flag.String("sourceType", "commerce", "The type of the event source.")
	sourceEnvironment := flag.String("sourceEnvironment", "stage", "The name of the event source environment.")

	flag.Parse()

	return &options{
		externalAPIPort:   *externalAPIPort,
		proxyPort:         *proxyPort,
		eventsTargetURL:   *eventsTargetURL,
		remoteEnvironment: *remoteEnvironment,
		namespace:         *namespace,
		requestTimeout:    *requestTimeout,
		skipVerify:        *skipVerify,
		proxyTimeout:      *proxyTimeout,
		requestLogging:    *requestLogging,
		proxyCacheTTL:     *proxyCacheTTL,
		sourceNamespace:   *sourceNamespace,
		sourceType:        *sourceType,
		sourceEnvironment: *sourceEnvironment,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --eventsTargetURL=%s"+
		" --remoteEnvironment=%s --namespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --sourceNamespace=%s --sourceType=%s --sourceEnvironment=%s --requestLogging=%t --proxyCacheTTL=%d",
		o.externalAPIPort, o.proxyPort, o.eventsTargetURL,
		o.remoteEnvironment, o.namespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.sourceNamespace, o.sourceType, o.sourceEnvironment, o.requestLogging, o.proxyCacheTTL)
}
