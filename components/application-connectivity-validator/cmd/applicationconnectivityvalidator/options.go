package main

import (
	"flag"
	"fmt"
)

type options struct {
	proxyPort              int
	tenant                 string
	group                  string
	eventServicePathPrefix string
	eventServiceHost       string
	appRegistryPathPrefix  string
	appRegistryHost        string
}

func parseArgs() *options {
	proxyPort := flag.Int("proxyPort", 8080, "Proxy port.")
	tenant := flag.String("tenant", "", "Name of the application tenant")
	group := flag.String("group", "", "Name of the application group")
	eventServicePathPrefix := flag.String("eventServicePathPrefix", "/v1/events", "Prefix of paths that will be directed to the Event Service")
	eventServiceHost := flag.String("eventServiceHost", "events-api:8080", "Host (and port) of the Event Service")
	appRegistryPathPrefix := flag.String("appRegistryPathPrefix", "/v1/metadata", "Prefix of paths that will be directed to the Application Registry")
	appRegistryHost := flag.String("appRegistryHost", "application-registry-external-api:8081", "Host (and port) of the Application Registry")

	flag.Parse()

	return &options{
		proxyPort: *proxyPort,
		tenant:    *tenant,
		group:     *group,
		eventServicePathPrefix: *eventServicePathPrefix,
		eventServiceHost:       *eventServiceHost,
		appRegistryPathPrefix:  *appRegistryPathPrefix,
		appRegistryHost:        *appRegistryHost,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--proxyPort=%d --tenant=%s --group=%s "+
		"--eventServicePathPrefix=%s --eventServiceHost=%s "+
		"--appRegistryPathPrefix=%s --appRegistryHost=%s",
		o.proxyPort, o.tenant, o.group,
		o.eventServicePathPrefix, o.eventServiceHost,
		o.appRegistryPathPrefix, o.appRegistryHost)
}
