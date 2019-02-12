package main

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	appName                       string
	externalAPIPort               int
	internalAPIPort               int
	namespace                     string
	tokenLength                   int
	appTokenExpirationMinutes     int
	runtimeTokenExpirationMinutes int
	caSecretName                  string
	requestLogging                bool
	connectorServiceHost          string
	applicationConnectorHost      string
	appRegistryHost               string
	eventsHost                    string
	appsInfoURL                   string
	runtimesInfoURL               string
	group                         string
	tenant                        string
}

type environment struct {
	country            string
	organization       string
	organizationalUnit string
	locality           string
	province           string
}

func parseArgs() *options {
	appName := flag.String("appName", "connector-service", "Name of the Certificate Service, used by k8s deployments and services.")
	externalAPIPort := flag.Int("externalAPIPort", 8081, "External API port.")
	internalAPIPort := flag.Int("internalAPIPort", 8080, "Internal API port.")
	namespace := flag.String("namespace", "kyma-integration", "Namespace used by Certificate Service")
	tokenLength := flag.Int("tokenLength", 64, "Length of a registration tokens.")
	appTokenExpirationMinutes := flag.Int("appTokenExpirationMinutes", 5, "Time to Live of application tokens expressed in minutes.")
	runtimeTokenExpirationMinutes := flag.Int("runtimeTokenExpirationMinutes", 10, "Time to Live of runtime tokens expressed in minutes.")
	caSecretName := flag.String("caSecretName", "nginx-auth-ca", "Name of the secret which contains root CA.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	connectorServiceHost := flag.String("connectorServiceHost", "cert-service.wormhole.cluster.kyma.cx", "Host at which this service is accessible.")
	// Temporary solution as only gateway.domain,name host is secured with client certificate. We should decide if we want to expose whole Connector Service through it.
	applicationConnectorHost := flag.String("applicationConnectorHost", "gateway.wormhole.cluster.kyma.cx", "Host secured with client certificate, used for certificate renewal.")
	appRegistryHost := flag.String("appRegistryHost", "", "Host at which this Application Registry is accessible.")
	eventsHost := flag.String("eventsHost", "", "Host at which this Event Service is accessible.")
	appsInfoURL := flag.String("appsInfoURL", "", "URL at which management information is available.")
	runtimesInfoURL := flag.String("runtimesInfoURL", "", "URL at which management information is available.")
	group := flag.String("group", "", "Default group")
	tenant := flag.String("tenant", "", "Default tenant")

	flag.Parse()

	return &options{
		appName:                       *appName,
		externalAPIPort:               *externalAPIPort,
		internalAPIPort:               *internalAPIPort,
		namespace:                     *namespace,
		tokenLength:                   *tokenLength,
		appTokenExpirationMinutes:     *appTokenExpirationMinutes,
		runtimeTokenExpirationMinutes: *runtimeTokenExpirationMinutes,
		caSecretName:                  *caSecretName,
		requestLogging:                *requestLogging,
		connectorServiceHost:          *connectorServiceHost,
		applicationConnectorHost:      *applicationConnectorHost,
		group:           *group,
		tenant:          *tenant,
		appRegistryHost: *appRegistryHost,
		eventsHost:      *eventsHost,
		appsInfoURL:                   *appsInfoURL,
		runtimesInfoURL:               *runtimesInfoURL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --externalAPIPort=%d --internalAPIPort=%d --namespace=%s --tokenLength=%d "+
		"--appTokenExpirationMinutes=%d --runtimeTokenExpirationMinutes=%d --caSecretName=%s --requestLogging=%t "+
		"--connectorServiceHost=%s --applicationConnectorHost=%s --appRegistryHost=%s --eventsHost=%s "+
		"--appsInfoURL=%s --runtimesInfoURL=%s --group=%s --tenant=%s",
		o.appName, o.externalAPIPort, o.internalAPIPort, o.namespace, o.tokenLength,
		o.appTokenExpirationMinutes, o.runtimeTokenExpirationMinutes, o.caSecretName, o.requestLogging,
		o.connectorServiceHost, o.applicationConnectorHost, o.appRegistryHost, o.eventsHost,
		o.appsInfoURL, o.runtimesInfoURL, o.group, o.tenant)
}

func parseEnv() *environment {
	return &environment{
		country:            os.Getenv("COUNTRY"),
		organization:       os.Getenv("ORGANIZATION"),
		organizationalUnit: os.Getenv("ORGANIZATIONALUNIT"),
		locality:           os.Getenv("LOCALITY"),
		province:           os.Getenv("PROVINCE"),
	}
}

func (e *environment) String() string {
	return fmt.Sprintf("COUNTRY=%s ORGANIZATION=%s ORGANIZATIONALUNIT=%s LOCALITY=%s PROVINCE=%s",
		e.country, e.organization, e.organizationalUnit, e.locality, e.province)
}
