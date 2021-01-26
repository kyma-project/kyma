package main

import (
	"flag"
	"fmt"

	"github.com/vrischmann/envconfig"
)

type args struct {
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
	profile                               string
	isBEBEnabled                          bool
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
	appName := flag.String("appName", "application-operator", "Name used in application controller registration")
	domainName := flag.String("domainName", "kyma.local", "Domain name of the cluster")
	namespace := flag.String("namespace", "kyma-integration", "Namespace in which the Application chart will be installed")
	syncPeriod := flag.Int("syncPeriod", 30, "Time period between resyncing existing resources")
	installationTimeout := flag.Int64("installationTimeout", 240, "Time after the release installation will time out")

	helmDriver := flag.String("helmDriver", "secret", "Backend storage driver used by Helm 3 to store release data")
	applicationGatewayImage := flag.String("applicationGatewayImage", "", "The image of the Application Gateway to use")
	applicationGatewayTestsImage := flag.String("applicationGatewayTestsImage", "", "The image of the Application Gateway Tests to use")
	eventServiceImage := flag.String("eventServiceImage", "", "The image of the Event Service to use")
	eventServiceTestsImage := flag.String("eventServiceTestsImage", "", "The image of the Event Service Tests to use")
	applicationConnectivityValidatorImage := flag.String("applicationConnectivityValidatorImage", "", "The image of the Application Connectivity Validator to use")
	gatewayOncePerNamespace := flag.Bool("gatewayOncePerNamespace", false, "Specifies if Gateway should be deployed once per Namespace based on ServiceInstance or for every Application")
	strictMode := flag.String("strictMode", "disabled", "Toggles Istio authorization policy for Validator and HTTP source adapter")
	healthPort := flag.String("healthPort", "8090", "Port for healthcheck server")
	profile := flag.String("profile", "", "Profile name")
	isBEBEnabled := flag.Bool("isBEBEnabled", false, "Toggles creation of eventing infrastructure based on BEB if BEB is enabled")

	flag.Parse()

	var c config
	if err := envconfig.InitWithPrefix(&c, "APP"); err != nil {
		return nil, err
	}

	return &options{
		args: args{
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
			profile:                               *profile,
			isBEBEnabled:                          *isBEBEnabled,
		},
		config: c,
	}, nil
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --domainName=%s --namespace=%s"+
		" --syncPeriod=%d --installationTimeout=%d --helmDriver=%s"+
		" --applicationGatewayImage=%s --applicationGatewayTestsImage=%s --eventServiceImage=%s --eventServiceTestsImage=%s"+
		" --applicationConnectivityValidatorImage=%s --gatewayOncePerNamespace=%v --strictMode=%s --healthPort=%s --profile=%s"+
		"--isBEBEnabled=%v APP_LOG_LEVEL=%s APP_LOG_FORMAT=%s",
		o.appName, o.domainName, o.namespace,
		o.syncPeriod, o.installationTimeout, o.helmDriver,
		o.applicationGatewayImage, o.applicationGatewayTestsImage, o.eventServiceImage, o.eventServiceTestsImage,
		o.applicationConnectivityValidatorImage, o.gatewayOncePerNamespace, o.strictMode, o.healthPort, o.profile,
		o.isBEBEnabled, o.LogLevel, o.LogFormat)
}
