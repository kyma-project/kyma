package main

import (
	"flag"
	"fmt"
)

type options struct {
	controllerSyncPeriod  int
	minimalConfigSyncTime int
	tenant                string
	runtimeId             string

	tokenURLConfigFile string
}

func parseArgs() *options {
	controllerSyncPeriod := flag.Int("controllerSyncPeriod", 60, "Time period between resyncing existing resources.")
	minimalConfigSyncTime := flag.Int("minimalConfigSyncTime", 300, "Minimal time between synchronizing configuration.")
	tenant := flag.String("tenant", "", "Tenant for whom runtime is provisioned.")
	runtimeId := flag.String("runtimeId", "", "ID of the Runtime.")

	tokenURLConfigFile := flag.String("tokenURLConfigFile", "/config/token", "File containing URL with token to initialize connection with Compass.")

	flag.Parse()

	return &options{
		controllerSyncPeriod:  *controllerSyncPeriod,
		minimalConfigSyncTime: *minimalConfigSyncTime,
		tenant:                *tenant,
		runtimeId:             *runtimeId,
		tokenURLConfigFile:    *tokenURLConfigFile,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--controllerSyncPeriod=%d --minimalConfigSyncTime=%d "+
		"--tenant=%s --runtimeId=%s --tokenURLConfigFile=%s",
		o.controllerSyncPeriod, o.minimalConfigSyncTime, o.tenant, o.runtimeId, o.tokenURLConfigFile)
}
