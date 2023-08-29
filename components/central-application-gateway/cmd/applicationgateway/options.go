package main

import (
	"flag"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type options struct {
	apiServerURL                string
	applicationSecretsNamespace string
	externalAPIPort             int
	kubeConfig                  string
	logLevel                    *zapcore.Level
	proxyCacheTTL               int
	proxyPort                   int
	proxyPortCompass            int
	proxyTimeout                int
	requestTimeout              int
}

func parseArgs(log *zap.Logger) (opts options) {
	flag.StringVar(&opts.apiServerURL, "apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&opts.applicationSecretsNamespace, "applicationSecretsNamespace", "kyma-system", "Namespace where Application secrets used by the Application Gateway exist")
	flag.IntVar(&opts.externalAPIPort, "externalAPIPort", 8081, "Port that exposes the API which allows checking the component status and exposes log configuration")
	flag.StringVar(&opts.kubeConfig, "kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster")
	opts.logLevel = zap.LevelFlag("logLevel", zap.InfoLevel, "Log level: panic | fatal | error | warn | info | debug. Can't be lower than info")
	flag.IntVar(&opts.proxyCacheTTL, "proxyCacheTTL", 120, "TTL, in seconds, for proxy cache of Remote API information")
	flag.IntVar(&opts.proxyPort, "proxyPort", 8080, "Port that acts as a proxy for the calls from services and Functions to an external solution in the default standalone mode or Compass bundles with a single API definition")
	flag.IntVar(&opts.proxyPortCompass, "proxyPortCompass", 8082, "Port that acts as a proxy for the calls from services and Functions to an external solution in the Compass mode")
	flag.IntVar(&opts.proxyTimeout, "proxyTimeout", 10, "Timeout for requests sent through the proxy, expressed in seconds")
	flag.IntVar(&opts.requestTimeout, "requestTimeout", 1, "Timeout for requests sent through Central Application Gateway, expressed in seconds")

	flag.Parse()

	opts.Log(log)

	return
}

func (o options) Log(log *zap.Logger) {
	log.Info("Parsed flags",
		zap.String("-apiServerURL", o.apiServerURL),
		zap.String("-applicationSecretsNamespace", o.applicationSecretsNamespace),
		zap.Int("-externalAPIPort", o.externalAPIPort),
		zap.String("-kubeConfig", o.kubeConfig),
		zap.String("-logLevel", o.logLevel.String()),
		zap.Int("-proxyCacheTTL", o.proxyCacheTTL),
		zap.Int("-proxyPort", o.proxyPort),
		zap.Int("-proxyPortCompass", o.proxyPortCompass),
		zap.Int("-proxyTimeout", o.proxyTimeout),
		zap.Int("-requestTimeout", o.requestTimeout),
	)
}
