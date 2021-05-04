package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort int
	proxyOSPort       int
	proxyMPSPort       int
	namespace       string
	requestTimeout  int
	skipVerify      bool
	proxyTimeout    int
	requestLogging  bool
	proxyCacheTTL   int
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyOSPort := flag.Int("proxyOSPort", 8080, "Proxy port for Kyma OS.")
	proxyMPSPort := flag.Int("proxyMPSPort", 8088, "Proxy port for Kyma MPS.")
	namespace := flag.String("namespace", "kyma-system", "Namespace used by the Application Gateway")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")

	flag.Parse()

	return &options{
		externalAPIPort:     *externalAPIPort,
		proxyOSPort:         *proxyOSPort,
		proxyMPSPort:        *proxyMPSPort,
		namespace:           *namespace,
		requestTimeout:      *requestTimeout,
		skipVerify:          *skipVerify,
		proxyTimeout:        *proxyTimeout,
		requestLogging:      *requestLogging,
		proxyCacheTTL:       *proxyCacheTTL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyOSPort=%d --proxyMPSPort=%d --namespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d",
		o.externalAPIPort, o.proxyOSPort, o.proxyMPSPort, o.namespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL)
}
