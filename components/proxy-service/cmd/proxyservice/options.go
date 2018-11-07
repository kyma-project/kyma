package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort   int
	proxyPort         int
	remoteEnvironment string
	namespace         string
	requestTimeout    int
	skipVerify        bool
	proxyTimeout      int
	requestLogging    bool
	proxyCacheTTL     int
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	remoteEnvironment := flag.String("remoteEnvironment", "default-ec", "Remote environment name for reading and updating services.")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")

	flag.Parse()

	return &options{
		externalAPIPort:   *externalAPIPort,
		proxyPort:         *proxyPort,
		remoteEnvironment: *remoteEnvironment,
		requestTimeout:    *requestTimeout,
		skipVerify:        *skipVerify,
		proxyTimeout:      *proxyTimeout,
		requestLogging:    *requestLogging,
		proxyCacheTTL:     *proxyCacheTTL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --remoteEnvironment=%s --namespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d",
		o.externalAPIPort, o.proxyPort, o.remoteEnvironment, o.namespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL)
}
