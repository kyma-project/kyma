package main

import (
	"flag"
	"fmt"
)

type options struct {
	controllerSyncPeriod   int
	minimalConfigFetchTime int

	tokenURLConfigFile string
}

func parseArgs() *options {
	controllerSyncPeriod := flag.Int("controllerSyncPeriod", 60, "Time period between resyncing existing resources.")
	minimalConfigFetchTime := flag.Int("minimalConfigFetchTime", 300, "Minimal time between fetching configuration.")

	tokenURLConfigFile := flag.String("tokenURLConfigFile", "/config/token", "File containing URL with token to initialize connection with Compass.")

	flag.Parse()

	return &options{
		controllerSyncPeriod:   *controllerSyncPeriod,
		minimalConfigFetchTime: *minimalConfigFetchTime,
		tokenURLConfigFile:     *tokenURLConfigFile,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--controllerSyncPeriod=%d --minimalConfigFetchTime=%d --tokenURLConfigFile=%s",
		o.controllerSyncPeriod, o.minimalConfigFetchTime, o.tokenURLConfigFile)
}
