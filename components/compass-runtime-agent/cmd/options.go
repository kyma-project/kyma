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
	namespace          string
	gatewayPort        int
}

func parseArgs() *options {
	controllerSyncPeriod := flag.Int("controllerSyncPeriod", 60, "Time period between resyncing existing resources.")
	minimalConfigSyncTime := flag.Int("minimalConfigSyncTime", 300, "Minimal time between synchronizing configuration.")
	tenant := flag.String("tenant", "", "Tenant for whom runtime is provisioned.")
	runtimeId := flag.String("runtimeId", "", "ID of the Runtime.")

	tokenURLConfigFile := flag.String("tokenURLConfigFile", "/config/token", "File containing URL with token to initialize connection with Compass.")
	namespace := flag.String("namespace", "kyma-integration", "Namespace the resources will be created in.")
	gatewayPort := flag.Int("gatewayPort", 8080, "Application Gateway port.")

	flag.Parse()

	return &options{
		controllerSyncPeriod:  *controllerSyncPeriod,
		minimalConfigSyncTime: *minimalConfigSyncTime,
		tenant:                *tenant,
		runtimeId:             *runtimeId,
		tokenURLConfigFile:    *tokenURLConfigFile,
		namespace:             *namespace,
		gatewayPort:           *gatewayPort,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--controllerSyncPeriod=%d --minimalConfigSyncTime=%d "+
		"--tenant=%s --runtimeId=%s --tokenURLConfigFile=%s --namespace=%s ==gatewayPort=%d",
		o.controllerSyncPeriod, o.minimalConfigSyncTime, o.tenant, o.runtimeId, o.tokenURLConfigFile, o.namespace, o.gatewayPort)
}
