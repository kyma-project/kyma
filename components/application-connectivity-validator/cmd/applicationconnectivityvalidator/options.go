package main

import (
	"flag"
	"fmt"
	"time"
)

type options struct {
	proxyPort                int
	externalAPIPort          int
	tenant                   string
	group                    string
	eventServicePathPrefixV1 string
	eventServicePathPrefixV2 string
	eventServiceHost         string
	eventMeshPathPrefix      string
	eventMeshHost            string
	eventMeshDestinationPath string
	appRegistryPathPrefix    string
	appRegistryHost          string
	appName                  string
	cacheExpirationMinutes   int
	cacheCleanupMinutes      int
	kubeConfig               string
	masterURL                string
	syncPeriod               time.Duration
}

func parseArgs() *options {
	proxyPort := flag.Int("proxyPort", 8081, "Proxy port.")
	externalAPIPort := flag.Int("externalAPIPort", 8080, "External API port.")
	tenant := flag.String("tenant", "", "Name of the application tenant")
	group := flag.String("group", "", "Name of the application group")
	eventServicePathPrefixV1 := flag.String("eventServicePathPrefixV1", "/v1/events", "Prefix of paths that will be directed to the Event Service V1")
	eventServicePathPrefixV2 := flag.String("eventServicePathPrefixV2", "/v2/events", "Prefix of paths that will be directed to the Event Service V2")
	eventServiceHost := flag.String("eventServiceHost", "events-api:8080", "Host (and port) of the Event Service")
	eventMeshPathPrefix := flag.String("eventMeshPathPrefix", "/events", "Prefix of paths that will be directed to the Event Mesh")
	eventMeshHost := flag.String("eventMeshHost", "events-adapter:8080", "Host (and port) of the Event Mesh adapter")
	eventMeshDestinationPath := flag.String("eventMeshDestinationPath", "/", "Path of the destination of the requests to the Event Mesh")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/v1/metadata", "Prefix of paths that will be directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")
	appName := flag.String("appName", "", "Name of the application CR the validator is created for")
	cacheExpirationMinutes := flag.Int("cacheExpirationMinutes", 1, "Expiration time for client IDs stored in cache expressed in minutes")
	cacheCleanupMinutes := flag.Int("cacheCleanupMinutes", 2, "Clean up time for client IDs stored in cache expressed in minutes")
	kubeConfig := flag.String("kubeConfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	masterURL := flag.String("masterURL", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	syncPeriod := flag.Duration("syncPeriod", 120*time.Second, "Sync period in seconds how often controller should periodically reconcile Application resource.")

	flag.Parse()

	return &options{
		proxyPort:                *proxyPort,
		externalAPIPort:          *externalAPIPort,
		tenant:                   *tenant,
		group:                    *group,
		eventServicePathPrefixV1: *eventServicePathPrefixV1,
		eventServicePathPrefixV2: *eventServicePathPrefixV2,
		eventServiceHost:         *eventServiceHost,
		eventMeshPathPrefix:      *eventMeshPathPrefix,
		eventMeshHost:            *eventMeshHost,
		eventMeshDestinationPath: *eventMeshDestinationPath,
		appRegistryPathPrefix:    *appRegistryPathPrefix,
		appRegistryHost:          *appRegistryHost,
		appName:                  *appName,
		cacheExpirationMinutes:   *cacheExpirationMinutes,
		cacheCleanupMinutes:      *cacheCleanupMinutes,
		kubeConfig:               *kubeConfig,
		masterURL:                *masterURL,
		syncPeriod:               *syncPeriod,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--proxyPort=%d --externalAPIPort=%d --tenant=%s --group=%s "+
		"--eventServicePathPrefixV1=%s --eventServicePathPrefixV2=%s --eventServiceHost=%s "+
		"--eventMeshPathPrefix=%s --eventMeshHost=%s "+
		"--eventMeshDestinationPath=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s --appName=%s "+
		"--cacheExpirationMinutes=%d --cacheCleanupMinutes=%d"+
		"--kubeConfig=%s --masterURL=%s --syncPeriod=%d",
		o.proxyPort, o.externalAPIPort, o.tenant, o.group,
		o.eventServicePathPrefixV1, o.eventServicePathPrefixV2, o.eventServiceHost,
		o.eventMeshPathPrefix, o.eventMeshHost, o.eventMeshDestinationPath,
		o.appRegistryPathPrefix, o.appRegistryHost, o.appName,
		o.cacheExpirationMinutes, o.cacheCleanupMinutes,
		o.kubeConfig, o.masterURL, o.syncPeriod)
}
