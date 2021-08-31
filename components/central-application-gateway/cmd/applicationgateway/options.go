package main

import (
	"flag"
	"fmt"
)

type options struct {
	externalAPIPort             int
	proxyPort                   int
	proxyPortCompass            int
	applicationSecretsNamespace string
	requestTimeout              int
	skipVerify                  bool
	proxyTimeout                int
	requestLogging              bool
	proxyCacheTTL               int
	kubeConfig                  string
	apiServerURL                string
}

func parseArgs() *options {
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port for Kyma OS or MPS bundles with a single API definition")
	proxyPortCompass := flag.Int("proxyPortCompass", 8082, "Proxy port for Kyma MPS.")
	applicationSecretsNamespace := flag.String("applicationSecretsNamespace", "kyma-integration", "Namespace where Application secrets used by the Application Gateway exist")
	requestTimeout := flag.Int("requestTimeout", 1, "Timeout for services.")
	skipVerify := flag.Bool("skipVerify", false, "Flag for skipping certificate verification for proxy target.")
	proxyTimeout := flag.Int("proxyTimeout", 10, "Timeout for proxy call.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	proxyCacheTTL := flag.Int("proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")
	kubeConfig := flag.String("kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	apiServerURL := flag.String("apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.Parse()

	return &options{
		externalAPIPort:             *externalAPIPort,
		proxyPort:                   *proxyPort,
		proxyPortCompass:            *proxyPortCompass,
		applicationSecretsNamespace: *applicationSecretsNamespace,
		requestTimeout:              *requestTimeout,
		skipVerify:                  *skipVerify,
		proxyTimeout:                *proxyTimeout,
		requestLogging:              *requestLogging,
		proxyCacheTTL:               *proxyCacheTTL,
		kubeConfig:                  *kubeConfig,
		apiServerURL:                *apiServerURL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --proxyPortCompass=%d --applicationSecretsNamespace=%s --requestTimeout=%d --skipVerify=%v --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d --kubeConfig=%s --apiServerURL=%s",
		o.externalAPIPort, o.proxyPort, o.proxyPortCompass, o.applicationSecretsNamespace, o.requestTimeout, o.skipVerify, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL, o.kubeConfig, o.apiServerURL)
}
