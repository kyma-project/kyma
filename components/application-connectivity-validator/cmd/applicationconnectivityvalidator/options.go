package main

import (
	"flag"
	"fmt"
)

type options struct {
	proxyPort                int
	externalAPIPort          int
	tenant                   string
	group                    string
	eventServicePathPrefixV1 string
	eventServicePathPrefixV2 string
	eventServiceHost         string
	appRegistryPathPrefix    string
	appRegistryHost          string
	cacheExpirationMinutes   int
	cacheCleanupMinutes      int
}

func parseArgs() *options {
	proxyPort := flag.Int("proxyPort", 8081, "Proxy port.")
	externalAPIPort := flag.Int("externalAPIPort", 8080, "External API port.")
	tenant := flag.String("tenant", "", "Name of the application tenant")
	group := flag.String("group", "", "Name of the application group")
	eventServicePathPrefixV1 := flag.String("eventServicePathPrefixV1", "/v1/events", "Prefix of paths that will be directed to the Event Service V1")
	eventServicePathPrefixV2 := flag.String("eventServicePathPrefixV2", "/v2/events", "Prefix of paths that will be directed to the Event Service V2")
	eventServiceHost := flag.String("eventServiceHost", "events-api:8080", "Host (and port) of the Event Service")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/v1/metadata", "Prefix of paths that will be directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")
	cacheExpirationMinutes := flag.Int("cacheExpirationMinutes", 1, "Expiration time for client IDs stored in cache expressed in minutes")
	cacheCleanupMinutes := flag.Int("cacheCleanupMinutes", 2, "Clean up time for client IDs stored in cache expressed in minutes")

	flag.Parse()

	return &options{
		proxyPort:                *proxyPort,
		externalAPIPort:          *externalAPIPort,
		tenant:                   *tenant,
		group:                    *group,
		eventServicePathPrefixV1: *eventServicePathPrefixV1,
		eventServicePathPrefixV2: *eventServicePathPrefixV2,
		eventServiceHost:         *eventServiceHost,
		appRegistryPathPrefix:    *appRegistryPathPrefix,
		appRegistryHost:          *appRegistryHost,
		cacheExpirationMinutes:   *cacheExpirationMinutes,
		cacheCleanupMinutes:      *cacheCleanupMinutes,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--proxyPort=%d --externalAPIPort=%d --tenant=%s --group=%s "+
		"--eventServicePathPrefixV1=%s --eventServicePathPrefixV2=%s --eventServiceHost=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s"+
		"--cacheExpirationMinutes=%d --cacheCleanupMinutes=%d",
		o.proxyPort, o.externalAPIPort, o.tenant, o.group,
		o.eventServicePathPrefixV1, o.eventServicePathPrefixV2, o.eventServiceHost,
		o.appRegistryPathPrefix, o.appRegistryHost,
		o.cacheExpirationMinutes, o.cacheCleanupMinutes)
}
