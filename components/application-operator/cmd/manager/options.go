package main

import (
	"flag"
	"fmt"
)

type options struct {
	appName                               string
	appMapName                            string
	domainName                            string
	namespace                             string
	syncPeriod                            int
	installationTimeout                   int64
	applicationGatewayImage               string
	applicationGatewayTestsImage          string
	eventServiceImage                     string
	eventServiceTestsImage                string
	applicationConnectivityValidatorImage string
	gatewayOncePerNamespace               bool
	strictMode                            string
	healthPort                            string
	helmDriver                            string
}

func parseArgs() *options {
	appName := flag.String("appName", "application-operator", "Name used in application controller registration")
	domainName := flag.String("domainName", "kyma.local", "Domain name of the cluster")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which the Application chart will be installed")
	syncPeriod := flag.Int("syncPeriod", 30, "Time period between resyncing existing resources")
	installationTimeout := flag.Int64("installationTimeout", 240, "Time after the release installation will time out")

	helmDriver := flag.String("helmDriver", "secret", "Toggles Helm 3 configuration storage between configMap and secret")
	applicationGatewayImage := flag.String("applicationGatewayImage", "", "The image of the Application Gateway to use")
	applicationGatewayTestsImage := flag.String("applicationGatewayTestsImage", "", "The image of the Application Gateway Tests to use")
	eventServiceImage := flag.String("eventServiceImage", "", "The image of the Event Service to use")
	eventServiceTestsImage := flag.String("eventServiceTestsImage", "", "The image of the Event Service Tests to use")
	applicationConnectivityValidatorImage := flag.String("applicationConnectivityValidatorImage", "", "The image of the Application Connectivity Validator to use")
	gatewayOncePerNamespace := flag.Bool("gatewayOncePerNamespace", false, "Specifies if Gateway should be deployed once per Namespace based on ServiceInstance or for every Application")
	strictMode := flag.String("strictMode", "disabled", "Toggles Istio authorization policy for Validator and HTTP source adapter")
	healthPort := flag.String("healthPort", "8090", "Port for healthcheck server")

	flag.Parse()

	return &options{
		appName:                               *appName,
		domainName:                            *domainName,
		namespace:                             *namespace,
		syncPeriod:                            *syncPeriod,
		installationTimeout:                   *installationTimeout,
		applicationGatewayImage:               *applicationGatewayImage,
		applicationGatewayTestsImage:          *applicationGatewayTestsImage,
		eventServiceImage:                     *eventServiceImage,
		eventServiceTestsImage:                *eventServiceTestsImage,
		applicationConnectivityValidatorImage: *applicationConnectivityValidatorImage,
		gatewayOncePerNamespace:               *gatewayOncePerNamespace,
		strictMode:                            *strictMode,
		healthPort:                            *healthPort,
		helmDriver:                            *helmDriver,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --domainName=%s --namespace=%s"+
		" --syncPeriod=%d --installationTimeout=%d --helmDriver=%s"+
		" --applicationGatewayImage=%s --applicationGatewayTestsImage=%s --eventServiceImage=%s --eventServiceTestsImage=%s"+
		" --applicationConnectivityValidatorImage=%s --gatewayOncePerNamespace=%v --strictMode=%s --healthPort=%s ",
		o.appName, o.domainName, o.namespace,
		o.syncPeriod, o.installationTimeout, o.helmDriver,
		o.applicationGatewayImage, o.applicationGatewayTestsImage, o.eventServiceImage, o.eventServiceTestsImage,
		o.applicationConnectivityValidatorImage, o.gatewayOncePerNamespace, o.strictMode, o.healthPort)
}
