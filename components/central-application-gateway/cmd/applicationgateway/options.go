package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
)

type options struct {
	apiServerURL                string
	applicationSecretsNamespace string
	externalAPIPort             int
	kubeConfig                  string
	logLevel                    logrus.Level
	proxyCacheTTL               int
	proxyPort                   int
	proxyPortCompass            int
	proxyTimeout                int
	requestLogging              bool
	requestTimeout              int
}

func parseArgs() (opts options) {
	flag.IntVar(&opts.externalAPIPort, "externalAPIPort", 8081, "External API port.")
	flag.IntVar(&opts.proxyPort, "proxyPort", 8080, "Proxy port for Kyma OS or MPS bundles with a single API definition")
	flag.IntVar(&opts.proxyPortCompass, "proxyPortCompass", 8082, "Proxy port for Kyma MPS.")
	flag.StringVar(&opts.applicationSecretsNamespace, "applicationSecretsNamespace", "kyma-system", "Namespace where Application secrets used by the Application Gateway exist")
	flag.IntVar(&opts.requestTimeout, "requestTimeout", 1, "Timeout for services.")
	flag.IntVar(&opts.proxyTimeout, "proxyTimeout", 10, "Timeout for proxy call.")
	flag.BoolVar(&opts.requestLogging, "requestLogging", false, "Flag for logging incoming requests.")
	flag.IntVar(&opts.proxyCacheTTL, "proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")
	flag.StringVar(&opts.kubeConfig, "kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&opts.apiServerURL, "apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	logLevel := flag.String("logLevel", "info", "Log level: panic | fatal | error | warn | info | debug | trace")

	flag.Parse()

	var err error
	opts.logLevel, err = logrus.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalln(err)
	}

	return
}

func (o options) String() string {
	return fmt.Sprintf("--externalAPIPort=%d --proxyPort=%d --proxyPortCompass=%d --applicationSecretsNamespace=%s --requestTimeout=%d --proxyTimeout=%d"+
		" --requestLogging=%t --proxyCacheTTL=%d --kubeConfig=%s --apiServerURL=%s --requestLogging=%s",
		o.externalAPIPort, o.proxyPort, o.proxyPortCompass, o.applicationSecretsNamespace, o.requestTimeout, o.proxyTimeout,
		o.requestLogging, o.proxyCacheTTL, o.kubeConfig, o.apiServerURL, o.logLevel)
}
