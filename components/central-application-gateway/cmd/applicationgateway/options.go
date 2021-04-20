package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	proxyPort       int
	namespace       string
	requestTimeout  int
	skipVerify      bool
	proxyTimeout    int
	requestLogging  bool
	proxyCacheTTL   int
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	namespace := flag.String("namespace", "kyma-system", "Namespace used by the Application Gateway")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")

	flag.Parse()

	return &options{
		externalAPIPort: *externalAPIPort,
		proxyPort:       *proxyPort,
		namespace:       *namespace,
		requestTimeout:  *requestTimeout,
		skipVerify:      *skipVerify,
		proxyTimeout:    *proxyTimeout,
		requestLogging:  *requestLogging,
		proxyCacheTTL:   *proxyCacheTTL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --namespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d",
		o.externalAPIPort, o.proxyPort, o.namespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL)
}
