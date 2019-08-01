package main

import (
	"flag"
	"fmt"
)

type options struct {
	controllerSyncPeriod  int
	minimalConfigSyncTime int
	integrationNamespace  string
	gatewayPort           int
}

func parseArgs() *options {
	controllerSyncPeriod := flag.Int("controllerSyncPeriod", 60, "Time period between resyncing existing resources.")
	minimalConfigSyncTime := flag.Int("minimalConfigSyncTime", 300, "Minimal time between synchronizing configuration.")
	integrationNamespace := flag.String("integrationNamespace", "kyma-integration", "Namespace the resources will be created in.")
	gatewayPort := flag.Int("gatewayPort", 8080, "Application Gateway port.")

	flag.Parse()

	return &options{
		controllerSyncPeriod:  *controllerSyncPeriod,
		minimalConfigSyncTime: *minimalConfigSyncTime,
		integrationNamespace:  *integrationNamespace,
		gatewayPort:           *gatewayPort,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--controllerSyncPeriod=%d --minimalConfigSyncTime=%d "+
		"--integrationNamespace=%s gatewayPort=%d",
		o.controllerSyncPeriod, o.minimalConfigSyncTime, o.integrationNamespace, o.gatewayPort)
}

type EnvConfig struct {
	DirectorURL string
	RuntimeId   string
	Tenant      string
}

func (ec EnvConfig) String() string {
	return fmt.Sprintf("DIRECTOR_URL=%s, RUNTIME_ID=%s, TENANT=%s", ec.DirectorURL, ec.RuntimeId, ec.Tenant)
}
