package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"github.com/sirupsen/logrus"
)

const (
	defaultCertificateValidityTime = 90 * 24 * time.Hour
	defaultNamespace               = "default"
)

type options struct {
	appName                        string
	externalAPIPort                int
	internalAPIPort                int
	namespace                      string
	tokenLength                    int
	appTokenExpirationMinutes      int
	runtimeTokenExpirationMinutes  int
	caSecretName                   types.NamespacedName
	rootCACertificateSecretName    types.NamespacedName
	requestLogging                 bool
	connectorServiceHost           string
	gatewayBaseURL                 string
	certificateProtectedHost       string
	appsInfoURL                    string
	appCertificateValidityTime     time.Duration
	runtimeCertificateValidityTime time.Duration
	central                        bool
	revocationConfigMapName        string
	lookupEnabled                  bool
	lookupConfigMapPath            string
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
	caSecretName := flag.String("caSecretName", "kyma-integration/nginx-auth-ca", "Namespace/name of the secret which contains certificate and key used for signing client certificates.")
	rootCACertificateSecretName := flag.String("rootCACertificateSecretName", "", "Namespace/name of the secret which contains root CA Certificate in case certificates are singed by intermediate CA.")
	requestLogging := flag.Bool("requestLogging", false, "Flag for logging incoming requests.")
	connectorServiceHost := flag.String("connectorServiceHost", "cert-service.wormhole.cluster.kyma.cx", "Host at which this service is accessible.")
	gatewayBaseURL := flag.String("gatewayBaseURL", "https://gateway.wormhole.cluster.kyma.cx", "Base URL of the gateway service.")
	certificateProtectedHost := flag.String("certificateProtectedHost", "gateway.wormhole.cluster.kyma.cx", "Host secured with client certificate, used for certificate renewal.")
	appsInfoURL := flag.String("appsInfoURL", "", "URL at which management information is available.")
	appCertificateValidityTime := flag.String("appCertificateValidityTime", "90d", "Validity time of certificates issued for apps by this service.")
	runtimeCertificateValidityTime := flag.String("runtimeCertificateValidityTime", "90d", "Validity time of certificates issued for runtimes by this service.")
	central := flag.Bool("central", false, "Determines whether connector works as the central")
	revocationConfigMapName := flag.String("revocationConfigMapName", "revocations-config", "Name of the config map containing revoked certificates")
	lookupEnabled := flag.Bool("lookupEnabled", false, "Determines whether connector should make a call to get gateway endpoint")
	lookupConfigMapPath := flag.String("lookupConfigMapPath", "/etc/config/config.json", "Path in the pod where Config Map for cluster lookup is stored")

	flag.Parse()

	appValidityTime, err := parseDuration(*appCertificateValidityTime)
	if err != nil {
		logrus.Infof("Failed to parse certificate validity time for applications: %s, using default value.", err)
	}

	runtimeValidityTime, err := parseDuration(*runtimeCertificateValidityTime)
	if err != nil {
		logrus.Infof("Failed to parse certificate validity time for applications: %s, using default value.", err)
	}

	return &options{
		appName:                        *appName,
		externalAPIPort:                *externalAPIPort,
		internalAPIPort:                *internalAPIPort,
		namespace:                      *namespace,
		tokenLength:                    *tokenLength,
		appTokenExpirationMinutes:      *appTokenExpirationMinutes,
		runtimeTokenExpirationMinutes:  *runtimeTokenExpirationMinutes,
		caSecretName:                   parseNamespacedName(*caSecretName),
		rootCACertificateSecretName:    parseNamespacedName(*rootCACertificateSecretName),
		requestLogging:                 *requestLogging,
		connectorServiceHost:           *connectorServiceHost,
		gatewayBaseURL:                 *gatewayBaseURL,
		certificateProtectedHost:       *certificateProtectedHost,
		central:                        *central,
		appsInfoURL:                    *appsInfoURL,
		appCertificateValidityTime:     appValidityTime,
		runtimeCertificateValidityTime: runtimeValidityTime,
		revocationConfigMapName:        *revocationConfigMapName,
		lookupEnabled:                  *lookupEnabled,
		lookupConfigMapPath:            *lookupConfigMapPath,
	}
}

func (o *options) String() string {
	return fmt.Sprintf("--appName=%s --externalAPIPort=%d --internalAPIPort=%d --namespace=%s --tokenLength=%d "+
		"--appTokenExpirationMinutes=%d --runtimeTokenExpirationMinutes=%d --caSecretName=%s --rootCACertificateSecretName=%s --requestLogging=%t "+
		"--connectorServiceHost=%s --certificateProtectedHost=%s --gatewayBaseURL=%s "+
		"--appsInfoURL=%s --central=%t --appCertificateValidityTime=%s --runtimeCertificateValidityTime=%s "+
		"--revocationConfigMapName=%s --lookupEnabled=%t --lookupConfigMapPath=%s",
		o.appName, o.externalAPIPort, o.internalAPIPort, o.namespace, o.tokenLength,
		o.appTokenExpirationMinutes, o.runtimeTokenExpirationMinutes, o.caSecretName, o.rootCACertificateSecretName, o.requestLogging,
		o.connectorServiceHost, o.certificateProtectedHost, o.gatewayBaseURL,
		o.appsInfoURL, o.central, o.appCertificateValidityTime, o.runtimeCertificateValidityTime,
		o.revocationConfigMapName, o.lookupEnabled, o.lookupConfigMapPath)
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
		return defaultCertificateValidityTime, fmt.Errorf("unrecognized time unit provided: %s", timeUnit)
	}

	timeLength, err := strconv.Atoi(durationString[:len(durationString)-1])
	if err != nil {
		return defaultCertificateValidityTime, err
	}

	return time.Duration(timeLength) * unitsMap[timeUnit], nil
}

func parseNamespacedName(value string) types.NamespacedName {
	parts := strings.Split(value, string(types.Separator))

	if singleValueProvided(parts) {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: defaultNamespace,
		}
	}

	namespace := get(parts, 0)
	if namespace == "" {
		namespace = defaultNamespace
	}

	return types.NamespacedName{
		Namespace: namespace,
		Name:      get(parts, 1),
	}
}

func singleValueProvided(split []string) bool {
	return len(split) == 1 || get(split, 1) == ""
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}
