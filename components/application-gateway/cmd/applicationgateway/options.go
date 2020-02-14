package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	proxyPort       int
	application     string
	namespace       string
	requestTimeout  int
	skipVerify      bool
	proxyTimeout    int
	requestLogging  bool
	proxyCacheTTL   int

	namespacedGateway bool
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	application := flag.String("application", "default-ec", "Application name for reading and updating services.")
	namespace := flag.String("namespace", "kyma-system", "Namespace used by the Application Gateway")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")
	namespacedGateway := flag.Bool("namespacedGateway", true, "Use Gateway in new configuration running as namespaced Gateway")

	flag.Parse()

	return &options{
		externalAPIPort: *externalAPIPort,
		proxyPort:       *proxyPort,
		application:     *application,
		namespace:       *namespace,
		requestTimeout:  *requestTimeout,
		skipVerify:      *skipVerify,
		proxyTimeout:    *proxyTimeout,
		requestLogging:  *requestLogging,
		proxyCacheTTL:   *proxyCacheTTL,

		namespacedGateway: *namespacedGateway,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --application=%s --namespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d --namespacedGateway=%t",
		o.externalAPIPort, o.proxyPort, o.application, o.namespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL, o.namespacedGateway)
}
