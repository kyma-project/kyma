package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const defaultCertificateValidityTime = 90 * 24 * time.Hour

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
	certificateProtectedHost      string
	appRegistryHost               string
	eventsHost                    string
	appsInfoURL                   string
	runtimesInfoURL               string
	group                         string
	tenant                        string
	certificateValidityTime       time.Duration
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
	// Temporary solution as only gateway.domain,name host is secured with client certificate. We should decide if we want to expose whole Connector Service through Nginx.
	certificateProtectedHost := flag.String("certificateProtectedHost", "gateway.wormhole.cluster.kyma.cx", "Host secured with client certificate, used for certificate renewal.")
	appRegistryHost := flag.String("appRegistryHost", "", "Host at which this Application Registry is accessible.")
	eventsHost := flag.String("eventsHost", "", "Host at which this Event Service is accessible.")
	appsInfoURL := flag.String("appsInfoURL", "", "URL at which management information is available.")
	runtimesInfoURL := flag.String("runtimesInfoURL", "", "URL at which management information is available.")
	group := flag.String("group", "", "Default group")
	tenant := flag.String("tenant", "", "Default tenant")
	certificateValidityTime := flag.String("certificateValidityTime", "90d", "Validity time of certificates issued by this service.")

	flag.Parse()

	validityTime, err := parseDuration(*certificateValidityTime)
	if err != nil {
		logrus.Infof("Failed to parse certificate validity time: %s, using default value.", err)
	}

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
		certificateProtectedHost:      *certificateProtectedHost,
		group:           *group,
		tenant:          *tenant,
		appRegistryHost: *appRegistryHost,
		eventsHost:      *eventsHost,
		appsInfoURL:     *appsInfoURL,
		runtimesInfoURL: *runtimesInfoURL,
		certificateValidityTime:       validityTime,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --externalAPIPort=%d --internalAPIPort=%d --namespace=%s --tokenLength=%d "+
		"--appTokenExpirationMinutes=%d --runtimeTokenExpirationMinutes=%d --caSecretName=%s --requestLogging=%t "+
		"--connectorServiceHost=%s --certificateProtectedHost=%s --appRegistryHost=%s --eventsHost=%s "+
		"--appsInfoURL=%s --runtimesInfoURL=%s --group=%s --tenant=%s --certificateValidityTime=%s",
		o.appName, o.externalAPIPort, o.internalAPIPort, o.namespace, o.tokenLength,
		o.appTokenExpirationMinutes, o.runtimeTokenExpirationMinutes, o.caSecretName, o.requestLogging,
		o.connectorServiceHost, o.certificateProtectedHost, o.appRegistryHost, o.eventsHost,
		o.appsInfoURL, o.runtimesInfoURL, o.group, o.tenant, o.certificateValidityTime)
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

func parseDuration(durationString string) (time.Duration, error) {
	unitsMap := map[string]time.Duration{"m": time.Minute, "h": time.Hour, "d": 24 * time.Hour}

	timeUnit := durationString[len(durationString)-1:]
	_, ok := unitsMap[timeUnit]
	if !ok {
		return defaultCertificateValidityTime, errors.New(fmt.Sprintf("unrecognized time unit provided: %s", timeUnit))
	}

	timeLength, err := strconv.Atoi(durationString[:len(durationString)-1])
	if err != nil {
		return defaultCertificateValidityTime, err
	}

	return time.Duration(timeLength) * unitsMap[timeUnit], nil
}
