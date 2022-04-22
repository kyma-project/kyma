package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/vrischmann/envconfig"
)

type args struct {
	proxyPort                int
	externalAPIPort          int
	tenant                   string
	group                    string
	eventingPathPrefixV1     string
	eventingPathPrefixV2     string
	eventingPublisherHost    string
	eventingPathPrefixEvents string
	eventingDestinationPath  string
	appRegistryPathPrefix    string
	appRegistryHost          string
	appName                  string
	cacheExpirationMinutes   int
	cacheCleanupMinutes      int
	kubeConfig               string
	apiServerURL             string
	syncPeriod               time.Duration
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
	tenant := flag.String("tenant", "", "Name of the application tenant")
	group := flag.String("group", "", "Name of the application group")
	eventingPathPrefixV1 := flag.String("eventingPathPrefixV1", "/v1/events", "Prefix of paths that will be directed to Kyma Eventing V1")
	eventingPathPrefixV2 := flag.String("eventingPathPrefixV2", "/v2/events", "Prefix of paths that will be directed to Kyma Eventing V2")
	eventingPublisherHost := flag.String("eventingPublisherHost", "eventing-event-publisher-proxy.kyma-system", "Host (and port) of the Eventing Publisher")
	eventingDestinationPath := flag.String("eventingDestinationPath", "/publish", "Path of the destination of the requests to the Eventing")
	eventingPathPrefixEvents := flag.String("eventingPathPrefixEvents", "/events", "Prefix of paths that will be directed to the Eventing")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/v1/metadata", "Prefix of paths that will be directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")
	appName := flag.String("appName", "", "Name of the application CR the validator is created for")
	cacheExpirationMinutes := flag.Int("cacheExpirationMinutes", 1, "Expiration time for client IDs stored in cache expressed in minutes")
	cacheCleanupMinutes := flag.Int("cacheCleanupMinutes", 2, "Clean up time for client IDs stored in cache expressed in minutes")
	kubeConfig := flag.String("kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	apiServerURL := flag.String("apiServerURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	syncPeriod := flag.Duration("syncPeriod", 120*time.Second, "Sync period in seconds how often controller should periodically reconcile Application resource.")

	flag.Parse()

	var c config
	if err := envconfig.InitWithPrefix(&c, "APP"); err != nil {
		return nil, err
	}

	return &options{
		args: args{
			proxyPort:                *proxyPort,
			externalAPIPort:          *externalAPIPort,
			tenant:                   *tenant,
			group:                    *group,
			eventingPathPrefixV1:     *eventingPathPrefixV1,
			eventingPathPrefixV2:     *eventingPathPrefixV2,
			eventingPublisherHost:    *eventingPublisherHost,
			eventingPathPrefixEvents: *eventingPathPrefixEvents,
			eventingDestinationPath:  *eventingDestinationPath,
			appRegistryPathPrefix:    *appRegistryPathPrefix,
			appRegistryHost:          *appRegistryHost,
			appName:                  *appName,
			cacheExpirationMinutes:   *cacheExpirationMinutes,
			cacheCleanupMinutes:      *cacheCleanupMinutes,
			kubeConfig:               *kubeConfig,
			apiServerURL:             *apiServerURL,
			syncPeriod:               *syncPeriod,
		},
		config: c,
	}, nil
}

func (o *options) String() string {
	return fmt.Sprintf("--proxyPort=%d --externalAPIPort=%d --tenant=%s --group=%s "+
		"--eventingPathPrefixV1=%s --eventingPathPrefixV2=%s --eventingPublisherHost=%s "+
		"--eventingPathPrefixEvents=%s --eventingDestinationPath=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s --appName=%s "+
		"--cacheExpirationMinutes=%d --cacheCleanupMinutes=%d "+
		"--kubeConfig=%s --apiServerURL=%s --syncPeriod=%d "+
		"APP_LOG_FORMAT=%s APP_LOG_LEVEL=%s",
		o.proxyPort, o.externalAPIPort, o.tenant, o.group,
		o.eventingPathPrefixV1, o.eventingPathPrefixV2, o.eventingPublisherHost,
		o.eventingPathPrefixEvents, o.eventingDestinationPath,
		o.appRegistryPathPrefix, o.appRegistryHost, o.appName,
		o.cacheExpirationMinutes, o.cacheCleanupMinutes,
		o.kubeConfig, o.apiServerURL, o.syncPeriod,
		o.LogFormat, o.LogLevel)
}
