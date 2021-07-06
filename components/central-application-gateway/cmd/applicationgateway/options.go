package main

import (
	"flag"
	"fmt"
)

type options struct {
	disableLegacyConnectivity   bool
	externalAPIPort             int
	proxyPort                   int
	proxyPortCompass            int
	applicationSecretsNamespace string
	requestTimeout              int
	skipVerify                  bool
	proxyTimeout                int
	requestLogging              bool
	proxyCacheTTL               int
}

func parseArgs() *options {
	disableLegacyConnectivity := flag.Bool("disableLegacyConnectivity", false, "Flag determining what HTTP handler will be used")
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port for Kyma OS.")
	proxyPortCompass := flag.Int("proxyPortCompass", 8088, "Proxy port for Kyma MPS.")
	applicationSecretsNamespace := flag.String("applicationSecretsNamespace", "kyma-integration", "Namespace where Application secrets used by the Application Gateway exist")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")

	flag.Parse()

	return &options{
		disableLegacyConnectivity:   *disableLegacyConnectivity,
		externalAPIPort:             *externalAPIPort,
		proxyPort:                   *proxyPort,
		proxyPortCompass:            *proxyPortCompass,
		applicationSecretsNamespace: *applicationSecretsNamespace,
		requestTimeout:              *requestTimeout,
		skipVerify:                  *skipVerify,
		proxyTimeout:                *proxyTimeout,
		requestLogging:              *requestLogging,
		proxyCacheTTL:               *proxyCacheTTL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--disableLegacyConnectivity=%t --externalAPIPort=%d --proxyPort=%d --proxyPortCompass=%d --applicationSecretsNamespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d",
		o.disableLegacyConnectivity, o.externalAPIPort, o.proxyPort, o.proxyPortCompass, o.applicationSecretsNamespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL)
}
