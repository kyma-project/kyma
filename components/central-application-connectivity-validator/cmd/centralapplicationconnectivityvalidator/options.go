package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/vrischmann/envconfig"
)

type args struct {
	proxyPort                   int
	externalAPIPort             int
	eventingPathPrefixV1        string
	eventingPathPrefixV2        string
	eventingPublisherHost       string
	eventingPathPrefixEvents    string
	eventingDestinationPath     string
	appRegistryPathPrefix       string
	appRegistryHost             string
	appNamePlaceholder          string
	cacheExpirationSeconds      int
	cacheCleanupIntervalSeconds int
	kubeConfig                  string
	apiServerURL                string
	syncPeriod                  time.Duration
}

type config struct {
	LogFormat string `default:"json"`
	LogLevel  string `default:"warn"`
}

type options struct {
	args
	config
}

func parseOptions() (*options, error) {
	proxyPort := flag.Int("proxyPort", 8081, "Proxy port.")
	externalAPIPort := flag.Int("externalAPIPort", 8080, "External API port.")
	eventingPathPrefixV1 := flag.String("eventingPathPrefixV1", "/v1/events", "Prefix of paths that is directed to Kyma Eventing V1")
	eventingPathPrefixV2 := flag.String("eventingPathPrefixV2", "/v2/events", "Prefix of paths that is directed to Kyma Eventing V2")
	eventingPublisherHost := flag.String("eventingPublisherHost", "eventing-event-publisher-proxy.kyma-system", "Host (and port) of the Eventing Publisher")
	eventingDestinationPath := flag.String("eventingDestinationPath", "/publish", "Path of the destination of the requests to the Eventing")
	eventingPathPrefixEvents := flag.String("eventingPathPrefixEvents", "/events", "Prefix of paths that is directed to the Cloud Events based Eventing")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/%%APP_NAME%%/v1/metadata", "Prefix of paths that is directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")
	appNamePlaceholder := flag.String("appNamePlaceholder", "%%APP_NAME%%", "Path URL placeholder used for an application name")
	cacheExpirationSeconds := flag.Int("cacheExpirationSeconds", 90, "Expiration time for client IDs stored in cache expressed in seconds")
	cacheCleanupIntervalSeconds := flag.Int("cacheCleanupIntervalSeconds", 15, "Clean up interval controls how often the client IDs stored in cache are removed")
	kubeConfig := flag.String("kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	apiServerURL := flag.String("apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	syncPeriod := flag.Duration("syncPeriod", 60*time.Second, "Sync period in seconds how often controller should periodically reconcile Application resource.")

	flag.Parse()

	var c config
	if err := envconfig.InitWithPrefix(&c, "APP"); err != nil {
		return nil, err
	}

	return &options{
		args: args{
			proxyPort:                   *proxyPort,
			externalAPIPort:             *externalAPIPort,
			eventingPathPrefixV1:        *eventingPathPrefixV1,
			eventingPathPrefixV2:        *eventingPathPrefixV2,
			eventingPublisherHost:       *eventingPublisherHost,
			eventingPathPrefixEvents:    *eventingPathPrefixEvents,
			eventingDestinationPath:     *eventingDestinationPath,
			appRegistryPathPrefix:       *appRegistryPathPrefix,
			appRegistryHost:             *appRegistryHost,
			appNamePlaceholder:          *appNamePlaceholder,
			cacheExpirationSeconds:      *cacheExpirationSeconds,
			cacheCleanupIntervalSeconds: *cacheCleanupIntervalSeconds,
			kubeConfig:                  *kubeConfig,
			apiServerURL:                *apiServerURL,
			syncPeriod:                  *syncPeriod,
		},
		config: c,
	}, nil
}

func (o *options) String() string {
	return fmt.Sprintf("--proxyPort=%d --externalAPIPort=%d "+
		"--eventingPathPrefixV1=%s --eventingPathPrefixV2=%s "+
		"--eventingPathPrefixEvents=%s --eventingPublisherHost=%s "+
		"--eventingDestinationPath=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s --appNamePlaceholder=%s "+
		"--cacheExpirationSeconds=%d --cacheCleanupIntervalSeconds=%d "+
		"--kubeConfig=%s --apiServerURL=%s --syncPeriod=%d "+
		"APP_LOG_FORMAT=%s APP_LOG_LEVEL=%s",
		o.proxyPort, o.externalAPIPort,
		o.eventingPathPrefixV1, o.eventingPathPrefixV2, o.eventingPathPrefixEvents,
		o.eventingPublisherHost, o.eventingDestinationPath,
		o.appRegistryPathPrefix, o.appRegistryHost, o.appNamePlaceholder,
		o.cacheExpirationSeconds, o.cacheCleanupIntervalSeconds,
		o.kubeConfig, o.apiServerURL, o.syncPeriod,
		o.LogFormat, o.LogLevel)
}

func (o *options) validate() error {
	if o.appNamePlaceholder == "" {
		return nil
	}
	if !strings.Contains(o.eventingPathPrefixV1, o.appNamePlaceholder) {
		return fmt.Errorf("eventingPathPrefixV1 '%s' should contain appNamePlaceholder '%s'", o.eventingPathPrefixV1, o.appNamePlaceholder)
	}
	if !strings.Contains(o.eventingPathPrefixV2, o.appNamePlaceholder) {
		return fmt.Errorf("eventingPathPrefixV2 '%s' should contain appNamePlaceholder '%s'", o.eventingPathPrefixV2, o.appNamePlaceholder)
	}
	if !strings.Contains(o.eventingPathPrefixEvents, o.appNamePlaceholder) {
		return fmt.Errorf("eventingPathPrefixEvents '%s' should contain appNamePlaceholder '%s'", o.eventingPathPrefixEvents, o.appNamePlaceholder)
	}
	if !strings.Contains(o.appRegistryPathPrefix, o.appNamePlaceholder) {
		return fmt.Errorf("appRegistryPathPrefix '%s' should contain appNamePlaceholder '%s'", o.appRegistryPathPrefix, o.appNamePlaceholder)
	}
	return nil
}
