package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/vrischmann/envconfig"
)

type args struct {
	proxyPort                int
	externalAPIPort          int
	eventServicePathPrefixV1 string
	eventServicePathPrefixV2 string
	eventServiceHost         string
	eventMeshPathPrefix      string
	eventMeshHost            string
	eventMeshDestinationPath string
	appRegistryPathPrefix    string
	appRegistryHost          string
	appNamePlaceholder       string
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
	eventServicePathPrefixV1 := flag.String("eventServicePathPrefixV1", "/%%APP_NAME%%/v1/events", "Prefix of paths that will be directed to the Event Service V1")
	eventServicePathPrefixV2 := flag.String("eventServicePathPrefixV2", "/%%APP_NAME%%/v2/events", "Prefix of paths that will be directed to the Event Service V2")
	eventServiceHost := flag.String("eventServiceHost", "events-api:8080", "Host (and port) of the Event Service")
	eventMeshPathPrefix := flag.String("eventMeshPathPrefix", "/%%APP_NAME%%/events", "Prefix of paths that will be directed to the Event Mesh")
	eventMeshHost := flag.String("eventMeshHost", "events-adapter:8080", "Host (and port) of the Event Mesh adapter")
	eventMeshDestinationPath := flag.String("eventMeshDestinationPath", "/", "Path of the destination of the requests to the Event Mesh")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/%%APP_NAME%%/v1/metadata", "Prefix of paths that will be directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")
	appNamePlaceholder := flag.String("appNamePlaceholder", "%%APP_NAME%%", "Path URL placeholder used for an application name")
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
			eventServicePathPrefixV1: *eventServicePathPrefixV1,
			eventServicePathPrefixV2: *eventServicePathPrefixV2,
			eventServiceHost:         *eventServiceHost,
			eventMeshPathPrefix:      *eventMeshPathPrefix,
			eventMeshHost:            *eventMeshHost,
			eventMeshDestinationPath: *eventMeshDestinationPath,
			appRegistryPathPrefix:    *appRegistryPathPrefix,
			appRegistryHost:          *appRegistryHost,
			appNamePlaceholder:       *appNamePlaceholder,
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
	return fmt.Sprintf("--proxyPort=%d --externalAPIPort=%d"+
		"--eventServicePathPrefixV1=%s --eventServicePathPrefixV2=%s --eventServiceHost=%s "+
		"--eventMeshPathPrefix=%s --eventMeshHost=%s "+
		"--eventMeshDestinationPath=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s --appNamePlaceholder=%s "+
		"--cacheExpirationMinutes=%d --cacheCleanupMinutes=%d "+
		"--kubeConfig=%s --apiServerURL=%s --syncPeriod=%d "+
		"APP_LOG_FORMAT=%s APP_LOG_LEVEL=%s",
		o.proxyPort, o.externalAPIPort,
		o.eventServicePathPrefixV1, o.eventServicePathPrefixV2, o.eventServiceHost,
		o.eventMeshPathPrefix, o.eventMeshHost, o.eventMeshDestinationPath,
		o.appRegistryPathPrefix, o.appRegistryHost, o.appNamePlaceholder,
		o.cacheExpirationMinutes, o.cacheCleanupMinutes,
		o.kubeConfig, o.apiServerURL, o.syncPeriod,
		o.LogFormat, o.LogLevel)
}

func (o *options) validate() error {
	if o.appNamePlaceholder == "" {
		return nil
	}
	if !strings.Contains(o.eventServicePathPrefixV1, o.appNamePlaceholder) {
		return fmt.Errorf("eventServicePathPrefixV1 '%s' should contain appNamePlaceholder '%s'", o.eventServicePathPrefixV1, o.appNamePlaceholder)
	}
	if !strings.Contains(o.eventServicePathPrefixV2, o.appNamePlaceholder) {
		return fmt.Errorf("eventServicePathPrefixV2 '%s' should contain appNamePlaceholder '%s'", o.eventServicePathPrefixV2, o.appNamePlaceholder)
	}
	if !strings.Contains(o.eventMeshPathPrefix, o.appNamePlaceholder) {
		return fmt.Errorf("eventMeshPathPrefix '%s' should contain appNamePlaceholder '%s'", o.eventMeshPathPrefix, o.appNamePlaceholder)
	}
	if !strings.Contains(o.appRegistryPathPrefix, o.appNamePlaceholder) {
		return fmt.Errorf("appRegistryPathPrefix '%s' should contain appNamePlaceholder '%s'", o.appRegistryPathPrefix, o.appNamePlaceholder)
	}
	return nil
}
