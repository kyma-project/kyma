package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	proxyPort       int
	minioURL        string
	namespace       string
	requestTimeout  int
	requestLogging  bool
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	minioURL := flag.String("minioURL", "localhost:9000", "Target URL for events to be sent.")
	namespace := flag.String("namespace", "kyma-system", "Namespace used by Gateway")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")

	flag.Parse()

	return &options{
		externalAPIPort: *externalAPIPort,
		proxyPort:       *proxyPort,
		minioURL:        *minioURL,
		namespace:       *namespace,
		requestTimeout:  *requestTimeout,
		requestLogging:  *requestLogging,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --minioURL=%s"+
		"--namespace=%s --requestTimeout=%d  --requestLogging=%t",
		o.externalAPIPort, o.proxyPort, o.minioURL,
		o.namespace, o.requestTimeout, o.requestLogging)
}
