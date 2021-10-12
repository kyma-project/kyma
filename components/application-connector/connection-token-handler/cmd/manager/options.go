package main

import (
	"flag"
	"fmt"
)

type options struct {
	syncPeriod          int
	connectorServiceURL string
	tokenTTL            int
}

func parseArgs() *options {
	connectorServiceURL := flag.String("connectorServiceURL", "http://connector-service-internal-api.kyma-integration.svc.cluster.local:8080", "connector-service internal URL")
	syncPeriod := flag.Int("syncPeriod", 30, "Time period (in seconds) between resyncing resources")
	tokenTTL := flag.Int("tokenTTL", 300, "Time period (in seconds) for token TTL")

	flag.Parse()

	return &options{
		syncPeriod:          *syncPeriod,
		connectorServiceURL: *connectorServiceURL,
		tokenTTL:            *tokenTTL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--syncPeriod=%d --connectorServiceURL=%s --tokenTTL=%d",
		o.syncPeriod, o.connectorServiceURL, o.tokenTTL)
}
