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
	domainName             string
	connectorServiceHost   string
	global                 bool
	externalKymaAPIPort    int
	internalKymaAPIPort    int
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
	domainName := flag.String("domainName", ".wormhole.cluster.kyma.cx", "Domain name used for URL generation.")
	connectorServiceHost := flag.String("connectorServiceHost", "cert-service.wormhole.cluster.kyma.cx", "Host at which this service is accessible.")
	global := flag.Bool("global", true, "Determines if Connector Service should act as global")
	externalKymaAPIPort := flag.Int("externalKymaAPIPort", 8083, "External Kyma Clusters API Port")
	internalKymaAPIPort := flag.Int("internalKymaAPIPort", 8082, "Internal Kyma Clusters API Port")

	flag.Parse()

	return &options{
		appName:                *appName,
		externalAPIPort:        *externalAPIPort,
		internalAPIPort:        *internalAPIPort,
		namespace:              *namespace,
		tokenLength:            *tokenLength,
		tokenExpirationMinutes: *tokenExpirationMinutes,
		domainName:             *domainName,
		connectorServiceHost:   *connectorServiceHost,
		global:                 *global,
		externalKymaAPIPort:    *externalKymaAPIPort,
		internalKymaAPIPort:    *internalKymaAPIPort,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --externalAPIPort=%d --internalAPIPort=%d --namespace=%s "+
		"--tokenLength=%d --tokenExpirationMinutes=%d --domainName=%s --connectorServiceHost=%s"+
		" --global=%t --externalKymaAPIPort=%d --internalKymaAPIPort=%d",
		o.appName, o.externalAPIPort, o.internalAPIPort, o.namespace,
		o.tokenLength, o.tokenExpirationMinutes, o.domainName, o.connectorServiceHost,
		o.global, o.externalKymaAPIPort, o.internalKymaAPIPort)
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
