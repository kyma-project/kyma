package main

import (
	"flag"
	"fmt"
	"os"
)

type options struct {
	appName                string
	externalAPIPort        int
	internalAPIPort        int
	namespace              string
	tokenLength            int
	tokenExpirationMinutes int
	caSecretName           string
	connectorServiceHost   string
	appRegistryHost        string
	eventsHost             string
	getInfoURL             string
	group                  string
	tenant                 string
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
	tokenExpirationMinutes := flag.Int("tokenExpirationMinutes", 5, "Time to Live of tokens expressed in minutes.")
	caSecretName := flag.String("caSecretName", "nginx-auth-ca", "Name of the secret which contains root CA.")
	connectorServiceHost := flag.String("connectorServiceHost", "cert-service.wormhole.cluster.kyma.cx", "Host at which this service is accessible.")
	appRegistryHost := flag.String("appRegistryHost", "", "Host at which this Application Registry is accessible.")
	eventsHost := flag.String("eventsHost", "", "Host at which this Event Service is accessible.")
	getInfoURL := flag.String("getInfoURL", "", "URL at which management information is available.")
	group := flag.String("group", "", "Default group")
	tenant := flag.String("tenant", "", "Default tenant")

	flag.Parse()

	return &options{
		appName:                *appName,
		externalAPIPort:        *externalAPIPort,
		internalAPIPort:        *internalAPIPort,
		namespace:              *namespace,
		tokenLength:            *tokenLength,
		tokenExpirationMinutes: *tokenExpirationMinutes,
		caSecretName:           *caSecretName,
		connectorServiceHost:   *connectorServiceHost,
		group:                  *group,
		tenant:                 *tenant,
		appRegistryHost:        *appRegistryHost,
		eventsHost:             *eventsHost,
		getInfoURL:             *getInfoURL,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --externalAPIPort=%d --internalAPIPort=%d --namespace=%s --tokenLength=%d "+
		"--tokenExpirationMinutes=%d --caSecretName=%s --connectorServiceHost=%s --appRegistryHost=%s --eventsHost=%s "+
		"--getInfoURL=%s --group=%s --tenant=%s",
		o.appName, o.externalAPIPort, o.internalAPIPort, o.namespace, o.tokenLength,
		o.tokenExpirationMinutes, o.caSecretName, o.connectorServiceHost, o.appRegistryHost, o.eventsHost,
		o.getInfoURL, o.group, o.tenant)
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
